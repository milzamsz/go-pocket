package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	adminEmail             = "admin@test.com"
	adminRequestedPassword = "admin"
	adminFallbackPassword  = "Admin#Test#2026!"
	userEmail              = "user@test.com"
	userAliasEmail         = "use@test.com"
	userRequestedPassword  = "user"
	userFallbackPassword   = "User#Test#2026!"
)

type seededCredential struct {
	Email        string
	Password     string
	UsedFallback bool
}

func RunLocalSeed() error {
	cfg := config.Load()
	if strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "production") {
		return errors.New("seed command is disabled in production environment")
	}

	pb := pocketbase.New()
	if err := Bootstrap(pb); err != nil {
		return fmt.Errorf("bootstrap app: %w", err)
	}
	if err := pb.Bootstrap(); err != nil {
		return fmt.Errorf("bootstrap pocketbase state: %w", err)
	}
	defer func() { _ = pb.ResetBootstrapState() }()
	if err := pb.RunAllMigrations(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	adminCreds, userCreds, userAliasCreds, err := seedLocalAccounts(context.Background(), pb)
	if err != nil {
		return err
	}

	fmt.Println("Local seed complete.")
	printCred("admin(superuser)", adminCreds)
	printCred("user(regular)", userCreds)
	printCred("user(alias)", userAliasCreds)
	return nil
}

func seedLocalAccounts(ctx context.Context, app core.App) (seededCredential, seededCredential, seededCredential, error) {
	adminCred, err := upsertSuperuser(ctx, app, adminEmail, adminRequestedPassword, adminFallbackPassword)
	if err != nil {
		return seededCredential{}, seededCredential{}, seededCredential{}, fmt.Errorf("seed admin account: %w", err)
	}
	userCred, err := upsertAuthUser(ctx, app, userEmail, userRequestedPassword, userFallbackPassword)
	if err != nil {
		return seededCredential{}, seededCredential{}, seededCredential{}, fmt.Errorf("seed user account: %w", err)
	}
	userAliasCred, err := upsertAuthUser(ctx, app, userAliasEmail, userRequestedPassword, userFallbackPassword)
	if err != nil {
		return seededCredential{}, seededCredential{}, seededCredential{}, fmt.Errorf("seed user alias account: %w", err)
	}

	// 1. Seed Default Org
	userRecord, err := app.FindAuthRecordByEmail("users", userEmail)
	if err != nil {
		return adminCred, userCred, userAliasCred, fmt.Errorf("find user for org seed: %w", err)
	}
	userAliasRecord, err := app.FindAuthRecordByEmail("users", userAliasEmail)
	if err != nil {
		return adminCred, userCred, userAliasCred, fmt.Errorf("find user alias for org seed: %w", err)
	}

	orgsColl, err := app.FindCollectionByNameOrId("organizations")
	if err != nil {
		return adminCred, userCred, userAliasCred, fmt.Errorf("find organizations: %w", err)
	}

	membersColl, err := app.FindCollectionByNameOrId("organization_members")
	if err != nil {
		return adminCred, userCred, userAliasCred, fmt.Errorf("find organization_members: %w", err)
	}

	orgRecord, err := app.FindFirstRecordByFilter("organizations", "slug = 'default'", nil)
	if err != nil {
		orgRecord = core.NewRecord(orgsColl)
		orgRecord.Set("id", "orgdefault12345")
		orgRecord.Set("slug", "default")
		orgRecord.Set("name", "Default Org")
		orgRecord.Set("owner", userRecord.Id)
		if err := app.SaveWithContext(ctx, orgRecord); err != nil {
			return adminCred, userCred, userAliasCred, fmt.Errorf("save default org: %w", err)
		}
	}

	memberships := []struct {
		UserID string
		Role   string
	}{
		{UserID: userRecord.Id, Role: "owner"},
		{UserID: userAliasRecord.Id, Role: "member"},
	}
	for _, entry := range memberships {
		_, findErr := app.FindFirstRecordByFilter("organization_members", "organization = {:orgID} && user = {:userID}", dbx.Params{"orgID": orgRecord.Id, "userID": entry.UserID})
		if findErr == nil {
			continue
		}
		memberRecord := core.NewRecord(membersColl)
		memberRecord.Set("organization", orgRecord.Id)
		memberRecord.Set("user", entry.UserID)
		memberRecord.Set("role", entry.Role)
		if saveErr := app.SaveWithContext(ctx, memberRecord); saveErr != nil {
			return adminCred, userCred, userAliasCred, fmt.Errorf("save default membership for %s: %w", entry.UserID, saveErr)
		}
	}

	// 2. Seed Products if empty
	prodColl, err := app.FindCollectionByNameOrId("products")
	if err == nil {
		count := 0
		_ = app.RecordQuery(prodColl).Select("count(*)").AndWhere(dbx.NewExp("organization = {:orgID}", dbx.Params{"orgID": orgRecord.Id})).Row(&count)
		if count == 0 {
			prods := []struct {
				Name     string
				Category string
				Price    int
				Stock    int
			}{
				{"Cyberpunk Processor", "electronics", 12900, 50},
				{"Techwear Jacket", "apparel", 8900, 20},
				{"Quantum Desk Lamp", "home", 4500, 15},
				{"Smart Hologram Cube", "electronics", 29900, 10},
			}
			for _, p := range prods {
				r := core.NewRecord(prodColl)
				r.Set("name", p.Name)
				r.Set("category", p.Category)
				r.Set("price", p.Price)
				r.Set("stock", p.Stock)
				r.Set("active", true)
				r.Set("organization", orgRecord.Id)
				_ = app.SaveWithContext(ctx, r)
			}
		}
	}

	// 3. Seed Kanban Columns & Cards if empty
	colColl, err := app.FindCollectionByNameOrId("kanban_columns")
	cardColl, err := app.FindCollectionByNameOrId("kanban_cards")
	if err == nil && cardColl != nil {
		colCount := 0
		_ = app.RecordQuery(colColl).Select("count(*)").AndWhere(dbx.NewExp("organization = {:orgID}", dbx.Params{"orgID": orgRecord.Id})).Row(&colCount)
		if colCount == 0 {
			columns := []string{"Todo", "In Progress", "Done"}
			colRecords := []*core.Record{}
			for idx, colName := range columns {
				r := core.NewRecord(colColl)
				r.Set("name", colName)
				r.Set("position", (idx+1)*1000)
				r.Set("organization", orgRecord.Id)
				if err := app.SaveWithContext(ctx, r); err == nil {
					colRecords = append(colRecords, r)
				}
			}

			if len(colRecords) == 3 {
				cards := []struct {
					Title       string
					Description string
					Badge       string
					ColIdx      int
				}{
					{"Design premium dark theme", "Create HSL tailored dark-mode tokens matching Stitch visual specification.", "High Priority", 0},
					{"Configure database migrations", "Create products and kanban board collections with strict tenant isolation.", "Tech Debt", 1},
					{"Implement tenant isolation", "Write multi-tenant middleware and service query filters.", "Completed", 2},
				}
				for idx, c := range cards {
					r := core.NewRecord(cardColl)
					r.Set("title", c.Title)
					r.Set("description", c.Description)
					r.Set("badge", c.Badge)
					r.Set("position", (idx+1)*1000)
					r.Set("column", colRecords[c.ColIdx].Id)
					r.Set("organization", orgRecord.Id)
					_ = app.SaveWithContext(ctx, r)
				}
			}
		}
	}

	return adminCred, userCred, userAliasCred, nil
}

func upsertSuperuser(ctx context.Context, app core.App, email, requestedPassword, fallbackPassword string) (seededCredential, error) {
	return upsertAuthRecord(ctx, app, core.CollectionNameSuperusers, email, requestedPassword, fallbackPassword, false)
}

func upsertAuthUser(ctx context.Context, app core.App, email, requestedPassword, fallbackPassword string) (seededCredential, error) {
	return upsertAuthRecord(ctx, app, "users", email, requestedPassword, fallbackPassword, true)
}

func upsertAuthRecord(
	ctx context.Context,
	app core.App,
	collectionName string,
	email string,
	requestedPassword string,
	fallbackPassword string,
	markVerified bool,
) (seededCredential, error) {
	record, err := app.FindAuthRecordByEmail(collectionName, email)
	if err != nil {
		collection, cErr := app.FindCollectionByNameOrId(collectionName)
		if cErr != nil {
			return seededCredential{}, fmt.Errorf("find %s collection: %w", collectionName, cErr)
		}
		record = core.NewRecord(collection)
		record.Set("email", email)
		if collectionName == "users" {
			record.Set("name", defaultNameFromEmail(email))
		}
	}

	tryPasswords := []string{requestedPassword}
	if strings.TrimSpace(fallbackPassword) != "" && fallbackPassword != requestedPassword {
		tryPasswords = append(tryPasswords, fallbackPassword)
	}
	for idx, candidate := range tryPasswords {
		record.Set("password", candidate)
		record.Set("passwordConfirm", candidate)
		if markVerified {
			record.Set("verified", true)
		}
		if saveErr := app.SaveWithContext(ctx, record); saveErr == nil {
			return seededCredential{
				Email:        email,
				Password:     candidate,
				UsedFallback: idx > 0,
			}, nil
		} else if idx == len(tryPasswords)-1 {
			return seededCredential{}, fmt.Errorf("save %s auth record: %w", email, saveErr)
		}
	}
	return seededCredential{}, errors.New("failed to seed auth record")
}

func defaultNameFromEmail(email string) string {
	localPart, _, found := strings.Cut(email, "@")
	if !found || strings.TrimSpace(localPart) == "" {
		return "User"
	}
	localPart = strings.TrimSpace(localPart)
	return strings.ToUpper(localPart[:1]) + localPart[1:]
}

func printCred(label string, cred seededCredential) {
	mode := "requested"
	if cred.UsedFallback {
		mode = "fallback"
	}
	fmt.Printf("- %s: email=%s password=%s (mode=%s)\n", label, cred.Email, cred.Password, mode)
}
