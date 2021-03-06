// SPDX-License-Identifier: MIT

package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/cnregion/id"
)

var data = []byte(`1:[2020,2019]:::1:2{33:浙江:1:0{}34:安徽:1:2{01:合肥:3:0{}02:芜湖:1:0{}}}`)

var obj = &DB{
	versions:          []int{2020, 2019},
	fullNameSeparator: "-",
	region: &Region{
		Name:      "",
		supported: 1,
		Items: []*Region{
			{
				ID:        "33",
				Name:      "浙江",
				supported: 1,
				FullName:  "浙江",
				FullID:    "330000000000",
				level:     id.Province,
			},
			{
				ID:        "34",
				Name:      "安徽",
				FullName:  "安徽",
				FullID:    "340000000000",
				supported: 1,
				level:     id.Province,
				Items: []*Region{
					{
						ID:        "01",
						Name:      "合肥",
						supported: 3,
						FullName:  "安徽-合肥",
						FullID:    "340100000000",
						level:     id.City,
					},
					{
						ID:        "02",
						Name:      "芜湖",
						supported: 1,
						FullName:  "安徽-芜湖",
						FullID:    "340200000000",
						level:     id.City,
					},
				},
			},
		},
	},
}

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	o1, err := Unmarshal(data, "-")
	a.NotError(err).
		Equal(o1.fullNameSeparator, obj.fullNameSeparator).
		Equal(o1.versions, obj.versions).
		Equal(len(o1.region.Items), len(obj.region.Items)).
		Equal(o1.region.Items[0].ID, obj.region.Items[0].ID).
		Equal(o1.region.Items[0].FullID, obj.region.Items[0].FullID).
		Equal(o1.region.Items[1].Items[0].ID, obj.region.Items[1].Items[0].ID).
		Equal(o1.region.Items[1].Items[0].FullID, obj.region.Items[1].Items[0].FullID)

	d1, err := obj.marshal()
	a.NotError(err).NotNil(d1)
	a.Equal(string(d1), string(data))

	_, err = Unmarshal([]byte("100:[2020]:::1:0{}"), "-")
	a.Equal(err, ErrIncompatible)
}

func TestDB_LoadDump(t *testing.T) {
	a := assert.New(t)

	path := filepath.Join(os.TempDir(), "cnregion_db.dict")
	a.NotError(obj.Dump(path, false))
	d, err := LoadFile(path, "-", false)
	a.NotError(err).NotNil(d)

	path = filepath.Join(os.TempDir(), "cnregion_db_compress.dict")
	a.NotError(obj.Dump(path, true))
	d, err = LoadFile(path, "-", true)
	a.NotError(err).NotNil(d)
}

func TestDB_Find(t *testing.T) {
	a := assert.New(t)

	r := obj.Find("34", "01")
	a.NotNil(r).Equal(r.Name, "合肥").Equal(r.FullName, "安徽-合肥")

	r = obj.Find("34", "01", "00")
	a.Nil(r)

	r = obj.Find("34")
	a.NotNil(r).Equal(r.Name, "安徽").Equal(r.FullName, "安徽")

	r = obj.Find()
	a.NotNil(r).Equal(r.Name, "").Equal(r.FullName, "").Equal(2, len(r.Items))

	// 不存在于 obj
	a.Nil(obj.Find("99"))
	a.Nil(obj.Find(""))
}

func TestDB_VersionIndex(t *testing.T) {
	a := assert.New(t)

	a.Equal(0, obj.VersionIndex(2020))
	a.Equal(1, obj.VersionIndex(2019))
	a.Equal(-1, obj.VersionIndex(1990))
}
