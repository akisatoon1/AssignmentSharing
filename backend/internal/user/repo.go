package user

type Repository interface {
	Save(usr User) error
}
