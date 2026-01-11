package models

// Permission represents who can see a secret entry
// If user_id = 0: entry is secret to everyone
// If user_id > 0: entry is personal secret, visible only to specified users
type Permission struct {
	Id     int64 `gorm:"column:id" json:"id"`          // References cl2003_msgs.id
	UserId int64 `gorm:"column:user_id" json:"user_id"` // Member number who can see (0 = secret to all)
}

func (Permission) TableName() string {
	return "cl2003_permissions"
}
