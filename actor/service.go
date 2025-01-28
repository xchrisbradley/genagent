package actor

import (
  "encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

// gonnna change the name because we will need to multiply this
type ActorID = uuid.UUID

var actordb = sqldb.NewDatabase("actor", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})


//encore:service
type Service struct{}
