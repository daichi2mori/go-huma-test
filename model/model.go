package model

type Options struct {
	Port int `doc:"Port to listen on." short:"p" default:"8888"`
}

type TodoResponse struct {
	ID          int64   `json:"id" example:"1" doc:"TodoのID"`
	Title       string  `json:"title" example:"買い物" doc:"Todoのタイトル"`
	Description *string `json:"description,omitempty" example:"牛乳を買う" doc:"Todoの詳細説明"`
	Completed   bool    `json:"completed" example:"false" doc:"完了状態"`
	CreatedAt   string  `json:"created_at" example:"2024-01-01T00:00:00Z" doc:"作成日時"`
	UpdatedAt   string  `json:"updated_at" example:"2024-01-01T00:00:00Z" doc:"更新日時"`
}

type ListTodosInput struct {
	Completed bool `query:"completed" doc:"完了状態でフィルタリング"`
}

type ListTodosOutput struct {
	Body struct {
		Todos []TodoResponse `json:"todos" doc:"Todoのリスト"`
	}
}

type GetTodoInput struct {
	ID int64 `path:"id" doc:"TodoのID"`
}

type GetTodoOutput struct {
	Body TodoResponse
}

type CreateTodoInput struct {
	Body struct {
		Title       string  `json:"title" minLength:"1" maxLength:"200" doc:"Todoのタイトル"`
		Description *string `json:"description,omitempty" maxLength:"1000" doc:"Todoの詳細説明"`
	}
}

type CreateTodoOutput struct {
	Body TodoResponse
}

type UpdateTodoInput struct {
	ID   int64 `path:"id" doc:"TodoのID"`
	Body struct {
		Title       string  `json:"title" minLength:"1" maxLength:"200" doc:"Todoのタイトル"`
		Description *string `json:"description,omitempty" maxLength:"1000" doc:"Todoの詳細説明"`
		Completed   bool    `json:"completed" doc:"完了状態"`
	}
}

type UpdateTodoOutput struct {
	Body TodoResponse
}

type DeleteTodoInput struct {
	ID int64 `path:"id" doc:"TodoのID"`
}

type DeleteTodoOutput struct {
	Body struct {
		Message string `json:"message" example:"Todo deleted successfully" doc:"削除結果メッセージ"`
	}
}

type ToggleTodoInput struct {
	ID int64 `path:"id" doc:"TodoのID"`
}

type ToggleTodoOutput struct {
	Body TodoResponse
}
