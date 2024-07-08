package models

//var ErrNoRecord = errors.New("models: подходящей записи не найдено")

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Tasks struct {
	Tasks []Task `json:"tasks"`
}

type User struct {
	Password string `json:"password"`
}
