package editor

// Cmd 返回当前编辑器命令
func Cmd(fileName string) string {
	return "subl -n -w " + fileName
}
