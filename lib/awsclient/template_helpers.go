package awsclient

type Rule struct {
	ToPort     int `json:",string"`
	FromPort   int `json:",string"`
	IpProtocol string
	CidrIp     interface{}
}

type Tag struct {
	Key   string
	Value string
}
