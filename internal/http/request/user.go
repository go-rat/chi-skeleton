package request

type UserID struct {
	ID uint `uri:"id" validate:"required,number"`
}

type AddUser struct {
	Name string `json:"name" form:"name" validate:"required,min=3,max=255" comment:"用户名"`
}

type UpdateUser struct {
	ID   uint   `uri:"id" validate:"required,number"`
	Name string `json:"name" form:"name" validate:"required,min=3,max=255" comment:"用户名"`
}
