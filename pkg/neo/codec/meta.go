// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

var (
	mMetaPrefix     = []byte("m_")
	namespacePrefix = []byte("n")
	databasePrefix  = []byte("d")
	matcherPrefix   = []byte("m")
	tablePrefix     = []byte("t")
	columnPrefix    = []byte("c")
	indexPrefix     = []byte("i")
	mMetaPrefixLen  = 3
)

// MetaKey
// key: m_n{namespace}
func MetaKey(namespace string) []byte {
	buf := make([]byte, 0, mMetaPrefixLen+len(namespace))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, namespacePrefix...)
	buf = append(buf, namespace...)
	return buf
}

//ColumnKey
// key: m_c{tableName}
func ColumnKey(columnName string) []byte {
	buf := make([]byte, 0, mMetaPrefixLen+len(columnName))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, columnPrefix...)
	buf = append(buf, columnName...)
	return buf
}

// TableKey
// key: m_t{tableName}
func TableKey(tableName string) []byte {
	buf := make([]byte, 0, mMetaPrefixLen+len(tableName))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, tablePrefix...)
	buf = append(buf, tableName...)
	return buf
}

// IndexKey
// key: m_i{indexName}
func IndexKey(indexName string) []byte {
	buf := make([]byte, 0, mMetaPrefixLen+len(indexName))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, indexPrefix...)
	buf = append(buf, indexName...)
	return buf
}

// MatcherKey
// key: m_m{matcher}
func MatcherKey(matcherName string) []byte {
	buf := make([]byte, 0, mMetaPrefixLen+len(matcherName))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, matcherPrefix...)
	buf = append(buf, matcherName...)
	return buf
}
