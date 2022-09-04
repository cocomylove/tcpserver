package iface

// 服务器运行的任务
type ITask interface {
	// 执行任务
	Run() error
	// 终止任务
	StopTask() error
	// 任务状态
	Status() bool
}
