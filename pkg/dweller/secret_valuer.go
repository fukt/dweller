package dweller

type SecretValuer interface {
	SecretValue(key string) (string, error)
}
