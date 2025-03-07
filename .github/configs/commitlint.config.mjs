import {
	RuleConfigCondition,
	RuleConfigSeverity,
	TargetCaseType,
} from '@commitlint/types';

export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    'header-max-length': [RuleConfigSeverity.Error, 'always', 128],
  },
};
