// Package logindex 提供面向日志行的倒排索引。
//
// 典型用法：Open → Add/Delete → Flush → Search(Term/And/Or/Not)。
// 写入在 Flush 后对查询可见；段合并后 posting 不丢不重；删除通过墓碑生效。
package logindex
