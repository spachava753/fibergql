package models

import "github.com/spachava753/fibergql/integration/remote_api"

type Viewer struct {
	User *remote_api.User
}
