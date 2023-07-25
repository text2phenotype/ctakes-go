package pipeline

type Pipeline func(request Request) <-chan string
