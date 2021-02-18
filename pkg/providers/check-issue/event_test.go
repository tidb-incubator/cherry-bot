package checkIssue

import (
	"fmt"
	"testing"
)

func TestFilterImg(t *testing.T) {
	//str1 := "sfdf"
	str := "352452aaaa 中文 <img >sss<><><><> <img width=\"956\" alt=\"屏幕快照 2021-02-03 下午8 13 16\" src=\"https://user-images.githubusercontent.com/5906259/106745942-c7cc2000-665c-11eb-9689-6bc8d77ce982.png\"> <><><>"
	fmt.Println(filterImg(str))
}

func TestFilterBractet(t *testing.T) {
	//str := "[中文]2525[]"
	str := "## Error Report\\r\\n\\r\\n**This repository is ONLY used to solve issues related to DOCS.\\r\\nFor other issues (related to TiDB, PD, etc), please move to [other repositories](https://github.com/pingcap/).**\\r\\n\\r\\nPlease answer the following questions before submitting your issue. Thanks!\\r\\n\\r\\n1. What is the URL/path of the document related to this issue?\\r\\n\\r\\nhttps://docs.pingcap.com/zh/tidb/dev/privilege-management\\r\\n\\r\\n2. How would you like to improve it?\\r\\n\\r\\ndelete grand in the table\\r\\n\\r\\n![企业微信截图_e7ea7242-875c-443f-9661-39bec203c1ee](https://user-images.githubusercontent.com/53471087/106870333-f228e680-670b-11eb-9048-14c1d5a729e7.png)\\r\\n\\r\\n\\r\\n\\r\\n\\r\\n"
	fmt.Println(filterBracket(str))
}
