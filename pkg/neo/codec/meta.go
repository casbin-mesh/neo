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
	mMetaPrefix     = []byte("m")
	mSchemaPrefix   = []byte("s")
	namespacePrefix = []byte("_n")
	matcherPrefix   = []byte("_m")
	tablePrefix     = []byte("_t")
	columnPrefix    = []byte("_c")
	indexPrefix     = []byte("_i")
	databasePrefix  = []byte("_d")
)

// MetaKey
// key: m_n{namespace}
func MetaKey(namespace string) []byte {
	buf := make([]byte, 0, len(mMetaPrefix)+len(namespacePrefix)+len(namespace))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, namespacePrefix...)
	buf = append(buf, namespace...)
	return buf
}

//ColumnKey
// key: m_t{tid}_c{tableName}
func ColumnKey(tid uint64, columnName string) []byte {
	buf := make([]byte, 0, len(columnName)+len(mMetaPrefix)+len(tablePrefix)+len(columnPrefix)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, tablePrefix...)
	buf = appendUint64(buf, tid)
	buf = append(buf, columnPrefix...)
	buf = append(buf, columnName...)
	return buf
}

// TableKey
// key: m_d{did}_t{tableName}
func TableKey(did uint64, tableName string) []byte {
	buf := make([]byte, 0, len(tableName)+len(mMetaPrefix)+len(databasePrefix)+len(tablePrefix)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, databasePrefix...)
	buf = appendUint64(buf, did)
	buf = append(buf, tablePrefix...)
	buf = append(buf, tableName...)
	return buf
}

// IndexKey
// key: m_t{tid}_i{indexName}
func IndexKey(tid uint64, indexName string) []byte {
	buf := make([]byte, 0, len(indexName)+len(mMetaPrefix)+len(tablePrefix)+len(indexPrefix)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, tablePrefix...)
	buf = appendUint64(buf, tid)
	buf = append(buf, indexPrefix...)
	buf = append(buf, indexName...)
	return buf
}

// MatcherKey
// key: m_d{did}_m{matcher}
func MatcherKey(did uint64, matcherName string) []byte {
	buf := make([]byte, 0, len(matcherName)+len(mMetaPrefix)+len(databasePrefix)+len(matcherPrefix)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, databasePrefix...)
	buf = appendUint64(buf, did)
	buf = append(buf, matcherPrefix...)
	buf = append(buf, matcherName...)
	return buf
}
