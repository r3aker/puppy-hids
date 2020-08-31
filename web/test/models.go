package main

type serverConfig struct {
	Learn bool `bson:"learn"`
	OfflineCheck bool `bson:"offlinecheck"`
	BlackList blackList `bson:",omitempty"`
	WhiteList whiteList `bson:",omitempty"`

	Private string `bson:"privatekey"`
	Public string `bson:"publickey"`
	Cert string `bson:"cert"`
}
type whiteList struct {
	File []string `bson:"file"`
	IP []string `bson:"ip"`
	Process []string `bson:"process"`
	Other []string `bson:"other"`
}
type blackList struct {
	File []string `bson:"file"`
	IP []string `bson:"ip"`
	Process []string `bson:"process"`
	Other []string `bson:"other"`
}


type rule struct {
	Type string `json:"type" bson:"type"` // 是否要regex 配置
	Data string `json:"data" bson:"data"`
}
type ruleInfo struct {
	Meta struct {
		Name        string `json:"name" bson:"name"`               // 名称
		Author      string `json:"author" bson:"author"`           // 编写人
		Description string `json:"description" bson:"description"` // 描述
		Level       int    `json:"level" bson:"level"`             // 风险等级
	} `json:"meta" bson:"meta"` // 规则信息
	Source string          `json:"source" bson:"source"` // 选择判断来源
	System string          `json:"system" bson:"system"` // 匹配系统
	Rules  map[string]rule `json:"rules" bson:"rules"`   // 具体匹配规则
	And    bool            `json:"and" bson:"and"`       // 规则逻辑
}

