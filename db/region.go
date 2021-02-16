// SPDX-License-Identifier: MIT

package db

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/issue9/errwrap"
)

// Region 表示单个区域
type Region struct {
	ID        string
	Name      string
	Supported int // 支持的版本号
	Items     []*Region
}

// IsSupported 当前数据是否支持该年份
func (reg *Region) IsSupported(db *DB, year int) bool {
	index := db.versionIndex(year)
	if index == -1 {
		return false
	}

	flag := 1 << index
	return reg.Supported&flag == flag
}

// AddItem 添加一条子项
func (reg *Region) AddItem(db *DB, id, name string, year int) error {
	index := db.versionIndex(year)
	if index == -1 {
		return fmt.Errorf("不支持该年份 %d 的数据", year)
	}

	for _, item := range reg.Items {
		if item.ID == id {
			return fmt.Errorf("已经存在相同 ID 的数据项：%s", id)
		}
	}

	reg.Items = append(reg.Items, &Region{
		ID:        id,
		Name:      name,
		Supported: 1 << index,
	})
	return nil
}

// SetSupported 设置当前数据支持指定的年份
func (reg *Region) SetSupported(db *DB, year int) error {
	index := db.versionIndex(year)
	if index == -1 {
		return fmt.Errorf("不存在该年份 %d 的数据", year)
	}

	flag := 1 << index
	if reg.Supported&flag == 0 {
		reg.Supported += flag
	}
	return nil
}

func (reg *Region) findItem(regionID ...string) *Region {
	if len(regionID) == 0 {
		return reg
	}

	for _, item := range reg.Items {
		if item.ID == regionID[0] {
			return item.findItem(regionID[1:]...)
		}
	}

	return nil
}

func (reg *Region) marshal(buf *errwrap.Buffer) error {
	buf.Printf("%s:%s:%d:%d{", reg.ID, reg.Name, reg.Supported, len(reg.Items))
	for _, item := range reg.Items {
		err := item.marshal(buf)
		if err != nil {
			return err
		}
	}
	buf.WByte('}')

	return nil
}

func (reg *Region) unmarshal(data []byte) error {
	index := indexByte(data, ':')
	reg.ID = string(data[:index])

	data = data[index+1:]
	index = indexByte(data, ':')
	reg.Name = string(data[:index])
	data = data[index+1:]

	index = indexByte(data, ':')
	supperted, err := strconv.Atoi(string(data[:index]))
	if err != nil {
		return err
	}
	reg.Supported = supperted
	data = data[index+1:]

	index = indexByte(data, '{')
	size, err := strconv.Atoi(string(data[:index]))
	if err != nil {
		return err
	}
	data = data[index+1:]

	if size > 0 {
		for i := 0; i < size; i++ {
			index := findEnd(data)
			if index < 0 {
				return errors.New("未找到结束符号 }")
			}

			item := &Region{}
			if err := item.unmarshal(data[:index]); err != nil {
				return err
			}
			reg.Items = append(reg.Items, item)
			data = data[index+1:]
		}
	}

	return nil
}

func indexByte(data []byte, b byte) int {
	index := bytes.IndexByte(data, b)
	if index == -1 {
		panic(fmt.Sprintf("在%s未找到：%s", string(data), string(b)))
	}
	return index
}

func findEnd(data []byte) int {
	deep := 0
	for i, b := range data {
		switch b {
		case '{':
			deep++
		case '}':
			deep--
			if deep == 0 {
				return i
			}
		}
	}

	return 0
}
