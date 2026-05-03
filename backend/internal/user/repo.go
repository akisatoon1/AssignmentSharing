package user

type Repository interface {
	Create(usr User) error
	Save(usr User) error
}
