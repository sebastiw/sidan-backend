package models

import (
	"fmt"
	"strings"
	"errors"
	"encoding/json"
)

type UserType string

const(
   MemberType UserType = "#"
   ProspectType UserType = "P"
   SuspectType UserType = "S"
)

//swagger:response User
type User struct {
	Type                         UserType `json:"type"`
	Number                       int64 `json:"number"`
	Email                        string `json:"email"`
	FulHaxPass                   string `json:"fulHaxPass"`
}

func (User) TableName() string {
  return "cl2007_members"
}

func (u *User) UnmarshalJSON(data []byte) error {
    // Define a secondary type so that we don't end up with a recursive call to json.Unmarshal
    type Aux User;
    var a *Aux = (*Aux)(u);
    err := json.Unmarshal(data, &a)
    if err != nil {
        return err
    }

    // Validate the valid enum values
    switch u.Type {
    case MemberType, ProspectType, SuspectType:
        return nil
    default:
        u.Type = ""
        return errors.New("invalid value for Key")
    }
}

func (u User) Fmt() string {
	s := make([]string, 0)
	s = addS(s, "Type", string(u.Type))
	s = addI(s, "Number", u.Number)
	s = addS(s, "Email", u.Email)
	return fmt.Sprintf("User{%s}", strings.Join(s, ", "))
}

