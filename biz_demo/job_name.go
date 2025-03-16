package biz_demo

const (
	Job1Name = "1"
	Job2Name = "2"
	Job3Name = "3"
	Job4Name = "4"
)

func init() {
	InitJobConstructorMap(map[string]JobConstructor{
		Job1Name: NewJob1,
		Job2Name: NewJob2,
		Job3Name: NewJob3,
		Job4Name: NewJob4,
	})
}
