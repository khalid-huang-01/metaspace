// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话状态定义
package common

const (
	ClientSessionStateReserved     = "RESERVED"  //创建好client session之后，为此状态
	ClientSessionStateTimeout      = "TIMEOUT"   //客户端与server session发起连接请求后，连接超时
	ClientSessionStateConnected    = "ACTIVE"    //client session与server session连接成功之后，活跃状态
	ClientSessionStateCompleted    = "COMPLETED" //一局游戏结束，client session与server session 断开连接
	ActivationClientSessionTimeout = 60          // client session 默认连接时间
)

const ClientSessionIDPrefix = "client-session-"
