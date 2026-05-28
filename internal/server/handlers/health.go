package handlers

import "github.com/pocketbase/pocketbase/core"

func Health(e *core.RequestEvent) error {
	return e.JSON(200, map[string]string{"status": "ok"})
}
