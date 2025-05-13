package enums

type PasswordPolicy int

const (
       PasswordPolicyNone   PasswordPolicy = iota // at least 1 char
       PasswordPolicyLow                          // at least 6 chars
       PasswordPolicyMedium                       // at least 8 chars. Must contain: 1 uppercase, 1 lowercase and 1 number
       PasswordPolicyHigh                         // at least 10 chars. Must contain: 1 uppercase, 1 lowercase, 1 number and 1 special character/symbol
)

func (p PasswordPolicy) String() string {
       return []string{"none", "low", "medium", "high"}[p]
}
