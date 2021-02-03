package checkIssue

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	//str1 := "sfdf"
	str := "352452aaaa 中文 <img >sss<><><><> <img width=\"956\" alt=\"屏幕快照 2021-02-03 下午8 13 16\" src=\"https://user-images.githubusercontent.com/5906259/106745942-c7cc2000-665c-11eb-9689-6bc8d77ce982.png\"> <><><>"
	fmt.Println(filterImg(str))
}
