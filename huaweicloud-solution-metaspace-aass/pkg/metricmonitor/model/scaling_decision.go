// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 伸缩决定模型
package model

const (
	ScalingDecisionActionIn   = "ScalingIn"
	ScalingDecisionActionOut  = "ScalingOut"
	ScalingDecisionActionNone = "NoScaling"
)

type ScalingDecision struct {
	// Action：伸缩动作：缩容/扩容/无伸缩
	Action string
	// CalculatedNum：根据指标计算出的伸缩数量
	CalculatedNum float64
	// AvailableNum：实例伸缩组内可伸缩的数量
	// 当实例伸缩组扩容时：可伸缩的数量 = 最大实例数 - 当前实例数
	// 当实例伸缩组缩容时：可伸缩的数量 = 当前实例数 - 最小实例数
	AvailableNum float64
	// ScalingNum：伸缩数量
	ScalingNum float64
	Instances  []string
}
