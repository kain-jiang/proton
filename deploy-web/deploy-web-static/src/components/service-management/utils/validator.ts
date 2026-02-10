// 校验状态
export const enum ValidateState {
  // 正常
  Normal,
  // 输入为空
  Empty,
}

export interface ServiceDeployValidateState {
  BatchJobName: ValidateState;
}

export const DefaultServiceDeployValidateState = {
  BatchJobName: ValidateState.Normal,
};
