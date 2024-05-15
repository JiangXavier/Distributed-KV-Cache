package service

// Picker 使用一致性哈希算法为key的查询请求找到应该发送到哪个节点
type Picker interface {
	Pick(key string) (Fetcher, bool)
}

// Fetcher is responsible for querying the value of key in the specified group cache.
// every distributed kv node should be implement this interface.
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

// Retriever 和后端数据库相连，“回调”
type Retriever interface {
	retrieve(string) ([]byte, error)
}

// 接口型函数

type RetrieveFunc func(key string) ([]byte, error)

func (f RetrieveFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}
