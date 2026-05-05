package service

// データを保存したり、取り出したりするため
type Repository interface {
	Create(usr User) error
	Save(usr User) error
	FindByID(id int64) (User, error)
}
