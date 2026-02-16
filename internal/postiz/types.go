package postiz

type Integration struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	Identifier string `json:"identifier"`
	Picture    string `json:"picture"`
	Disabled   bool   `json:"disabled"`
}

type IntegrationCheck struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

type Slot struct {
	Time string `json:"time"`
}

type Post struct {
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	PublishDate string   `json:"publishDate"`
	Integration []string `json:"integration"`
	Group       string   `json:"group"`
}

type CreatePostInput struct {
	Type        string   `json:"type"`
	Date        string   `json:"date,omitempty"`
	Content     string   `json:"content"`
	Integration []string `json:"integration"`
}

type Upload struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
