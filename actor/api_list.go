package actor

import (
  "encore.dev/types/uuid"
)

type ListActorRequest struct {
  IDs []uuid.UUID `json:"ids"`
}
