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
	mMetaPrefix   = []byte("m")
	mSchemaPrefix = []byte("s")
	tablePrefix   = []byte("t")

	namespacePrefixSep = []byte("_n")
	matcherPrefixSep   = []byte("_m")
	tablePrefixSep     = []byte("_t")
	tupleRecordPrefix  = []byte("_r")
	columnPrefixSep    = []byte("_c")
	indexPrefixSep     = []byte("_i")
	databasePrefixSep  = []byte("_d")
)

// MetaKey
// key: m_n{namespace}
func MetaKey(namespace string) []byte {
	buf := make([]byte, 0, len(mMetaPrefix)+len(namespacePrefixSep)+len(namespace))
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, namespacePrefixSep...)
	buf = append(buf, namespace...)
	return buf
}

//ColumnKey
// key: m_t{tid}_c{columnName}
func ColumnKey(tid uint64, columnName string) []byte {
	buf := make([]byte, 0, len(columnName)+len(mMetaPrefix)+len(tablePrefixSep)+len(columnPrefixSep)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, tablePrefixSep...)
	buf = appendUint64(buf, tid)
	buf = append(buf, columnPrefixSep...)
	buf = append(buf, columnName...)
	return buf
}

// TableKey
// key: m_d{did}_t{tableName}
func TableKey(did uint64, tableName string) []byte {
	buf := make([]byte, 0, len(tableName)+len(mMetaPrefix)+len(databasePrefixSep)+len(tablePrefixSep)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, databasePrefixSep...)
	buf = appendUint64(buf, did)
	buf = append(buf, tablePrefixSep...)
	buf = append(buf, tableName...)
	return buf
}

// IndexKey
// key: m_t{tid}_i{indexName}
func IndexKey(tid uint64, indexName string) []byte {
	buf := make([]byte, 0, len(indexName)+len(mMetaPrefix)+len(tablePrefixSep)+len(indexPrefixSep)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, tablePrefixSep...)
	buf = appendUint64(buf, tid)
	buf = append(buf, indexPrefixSep...)
	buf = append(buf, indexName...)
	return buf
}

// MatcherKey
// key: m_d{did}_m{matcher}
func MatcherKey(did uint64, matcherName string) []byte {
	buf := make([]byte, 0, len(matcherName)+len(mMetaPrefix)+len(databasePrefixSep)+len(matcherPrefixSep)+8)
	buf = append(buf, mMetaPrefix...)
	buf = append(buf, databasePrefixSep...)
	buf = appendUint64(buf, did)
	buf = append(buf, matcherPrefixSep...)
	buf = append(buf, matcherName...)
	return buf
}
