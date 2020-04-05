package models

import (
	"github.com/astaxie/beego/orm"
)

func GetHistoryLimit100() (his []*History, err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable("history").OrderBy("-updated").Limit(100).All(&his)

	return
}

func ReturnTimeIfExist(addr, ip string) (*History, error) {
	o := orm.NewOrm()
	cond := orm.NewCondition().Or("address", addr).Or("ip", ip)

	his := &History{}
	err := o.QueryTable("history").SetCond(cond).OrderBy("-updated").Limit(1).One(his)
	if err != nil {
		return nil, err
	}

	return his, nil
}
