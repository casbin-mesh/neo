package model

import (
	"errors"
	"strings"
)

var ErrTableNotExists = errors.New("table not exits")

type DBInfo struct {
	ID          uint64
	Name        CIStr
	TableInfo   []*TableInfo
	MatcherInfo []*MatcherInfo
}

func (d *DBInfo) Clone() *DBInfo {
	nd := *d
	nd.MatcherInfo = make([]*MatcherInfo, len(d.MatcherInfo))
	for i, info := range d.MatcherInfo {
		nd.MatcherInfo[i] = info.Clone()
	}
	nd.TableInfo = make([]*TableInfo, len(d.TableInfo))
	for i, info := range d.TableInfo {
		nd.TableInfo[i] = info.Clone()
	}
	return &nd
}

func (d *DBInfo) TableByName(name string) (*TableInfo, error) {
	for _, info := range d.TableInfo {
		if info.Name.L == strings.ToLower(name) {
			return info, nil
		}
	}
	return nil, ErrTableNotExists
}

func (d *DBInfo) TableById(id uint64) (*TableInfo, error) {
	for _, info := range d.TableInfo {
		if info.ID == id {
			return info, nil
		}
	}
	return nil, ErrTableNotExists
}
