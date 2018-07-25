package models

import (
	"github.com/astaxie/beego/orm"
)

func GetHistoryLimit100() (his []*History) {
	o := orm.NewOrm()
	o.QueryTable("history").OrderBy("-updated").Limit(100).All(&his)
	return
}

func ReturnTimeIfExist(addr, ip string) *History {
	o := orm.NewOrm()
	cond := orm.NewCondition()
	cond1 := cond.Or("address", addr).Or("ip", ip)

	his := &History{}
	err := o.QueryTable("history").SetCond(cond1).OrderBy("-updated").Limit(1).One(his)
	if err != nil {
		return nil
	}

	return his
}
