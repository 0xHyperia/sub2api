export default {
  distribution: {
    agentCustomers: '查看客户',
    agentViewDetails: '查看代理详情',
    agentEditRate: '调整返佣比例',
    agentPermissions: '管理权限',
    agentRewards: '配置拉新奖励',
    agentSuspend: '暂停代理',
    agentActivate: '恢复代理',
    agentRevoke: '撤销代理',
    filterCurrent: '当前筛选',
    clearFilters: '清除全部',
    anomalySeverityCritical: '严重',
    anomalySeverityHigh: '高风险',
    anomalySeverityMedium: '需关注',
    anomalyOverdueWithdrawal: '提现审核超时',
    anomalyPendingFx: '汇率待处理',
    anomalyAgentDebt: '代理钱包负债',
    anomalyInactiveCustomers: '非活跃代理客户',
    anomalyMaturedCommission: '佣金解冻延迟',
    analytics: {
      overviewTitle: '分销总览', overviewDescription: '分销渠道获客、转化、成交与佣金经营分析', channelTrend: '渠道经营趋势', channelComposition: '渠道业务构成', channelCompositionHint: '区分一级代理直属客户与二级代理团队业务', topAgents: '一级代理贡献排行', topAgentsHint: '按本期客户实付从高到低', allAgents: '全部代理', noAgentSales: '当前周期暂无代理成交', rankingCommission: '佣金 {amount}', liability: '佣金负债与结算', liabilityHint: '平台当前对代理的应付及风险构成', ledger: '查看台账', availableCommission: '可结算佣金', frozenCommission: '冻结佣金', reservedCommission: '提现占用', registrationRewards: '本期注册奖励', rechargeRewards: '本期充值奖励', paidThisMonth: '本月已付款', reversals: '累计冲正', agentDebt: '代理负债', operations: '运营待办', operationsHint: '优先处理影响代理结算的事项', reconciliation: '异常对账', pendingWithdrawals: '待审核提现', payingWithdrawals: '打款处理中', withdrawalTotal: '合计 {amount}', overdueReview: '审核超时', overdueHint: '超过 24 小时未完成审核', maturity: '佣金解冻', maturityRunning: '运行中', maturityHealthy: '正常', maturityStandby: '待机', maturityError: '异常', maturityDisabled: '未启用', maturityPending: '待运行', maturityReleased: '{time}，释放 {count} 条', maturityWaiting: '等待首次执行', loadFailed: '加载分销总览失败'
    },
    agentAnalytics: {
      emailMissing: '未设置邮箱', usernameMissing: '未设置用户名', l1: '一级代理', l2: '二级代理', periodRate: '本期新增 / 付费', periodPaid: '本期客户实付', periodCommission: '本期佣金', directTeam: '直属 {direct} · 团队 {team}', rate: '返佣比例', balance: '可用 / 冻结', details: '查看代理详情'
    },
    agents: {
      title: '代理管理', description: '管理一级代理及其团队关系', period: '经营周期', addL1: '添加一级代理', addL2: '添加二级代理', upgrade: '升级为代理', exportFiltered: '导出当前筛选', export: '导出', columnSettings: '列设置', resetColumns: '恢复默认列', searchPlaceholder: '搜索邮箱、用户名或推广码', search: '搜索代理', level: '代理等级', allLevels: '全部等级', status: '代理状态', allStatuses: '全部状态', active: '有效', suspended: '已暂停', revoked: '已撤销', currentFilters: '当前筛选', searchChip: '搜索：{query}', clearAll: '清除全部', platformDirect: '平台主管', parent: '上级：{email}', unknown: '未知', systemDefault: '系统默认', customRate: '自定义比例', payingCount: '付费 {count}', newPayingRate: '新增 / 付费 · {rate}', directTeam: '直属 {direct} · 团队 {team}', directTeamCounts: '直属 {directNew}/{directPaying} · 团队 {teamNew}/{teamPaying}', frozenAmount: '冻结 {amount}', viewDetails: '查看详情', noMatch: '暂无符合条件的代理', noMatchHint: '调整筛选条件，或添加第一个一级代理',
      stepLabel: '添加代理步骤', stepUser: '选择用户', stepRelation: '代理关系', stepBusiness: '经营配置', selectUser: '选择用户', userPlaceholder: '输入用户邮箱或用户名', selectUserHint: '先选择现有用户。已绑定代理的客户也可以升级，历史归属不会被重写。', parentL1: '所属一级代理', selectParent: '选择一级代理', upgradeHint: '该用户当前绑定了分销代理，确认后将升级为{level}代理；历史订单和佣金归属不变。', rate: '返佣比例', defaultRateHint: '留空使用系统默认比例 {rate}%', advanced: '高级设置', customCode: '自定义推广码', autoCode: '留空由系统自动生成', confirmConfig: '确认代理配置', user: '用户', promotionCode: '推广码', autoGenerate: '自动生成', cancel: '取消', previous: '上一步', next: '下一步', adding: '添加中...', confirmAdd: '确认添加',
      detailTitle: '代理详情', detailDescription: '经营数据、权限、奖励和审计记录', detailPeriod: '详情经营周期', tabOverview: '经营概览', tabCustomers: '客户与团队', tabCommission: '佣金与结算', tabStrategy: '策略与权限', tabAudit: '审计记录', directBusiness: '直属业务', teamBusiness: '团队业务', newPaying: '新增 / 付费', conversion: '转化率', customerPaid: '客户实付', directCommission: '直属佣金', activeChildren: '活跃下级', teamPaid: '团队实付', teamCommission: '团队佣金', childRanking: '下级代理贡献排行', acquiredPaid: '{newCount} 新增 · {payingCount} 付费', availableCommission: '可结算佣金', frozenCommission: '冻结佣金', reservedCommission: '提现占用', refundDebt: '退款负债', periodCommission: '本期佣金', cumulativeCommission: '累计佣金', cumulativeWithdrawn: '累计提现', convertedBalance: '转为余额', recruitment: '下级招募', promotionStats: '推广统计', authorized: '已授权', unauthorized: '未授权', viewable: '可查看', noPermission: '无权限', changeLog: '变更记录', recentEvents: '最近 {count} 条', system: '系统', noNote: '无备注', noEvents: '暂无变更记录', closeStats: '关闭统计权限', openStats: '开通统计权限', revokeRecruitment: '撤销招募权限', grantRecruitment: '授予招募权限', editRate: '编辑返佣比例', suspendAgent: '暂停代理', reactivateAgent: '重新启用', restoreAgent: '恢复代理', revokeAgent: '撤销代理',
      statsCloseTitle: '关闭推广统计权限', statsOpenTitle: '开通推广统计权限', statsCloseHint: '关闭后代理仍可使用推广链接，但不能查看访问和转化数据。', statsOpenHint: '开通后代理可以查看自己推广链接的匿名访问和注册转化数据。', operationReason: '操作原因', statsReasonPlaceholder: '记录开通或关闭依据', saving: '保存中...', confirm: '确认', revokeRecruitmentHint: '撤销后不能继续添加下级代理，已有团队关系不受影响。', grantRecruitmentHint: '授权后，该一级代理可以通过完整邮箱添加下级代理。', recruitmentReasonPlaceholder: '记录授权或撤销依据', currentEffective: '当前生效 {rate} · {mode}', inheritDefault: '沿用系统默认', ratePolicy: '返佣策略', changeReason: '变更原因', changeReasonPlaceholder: '说明本次比例调整依据', auditHint: '变更前后比例和操作人将写入审计记录', defaultRateValue: '系统默认值 {rate}%', saveRate: '保存比例', statusReasonPlaceholder: '记录本次状态调整原因', processing: '处理中...',
      statusReenableTitle: '重新启用代理', statusRestoreTitle: '恢复代理', statusSuspendTitle: '暂停代理', statusRevokeTitle: '撤销代理', revokeMessage: '撤销后该代理不会再产生新的客户和佣金关系，管理员之后仍可重新启用。', suspendMessage: '暂停后将停止该代理的新客户绑定和新佣金产生。', reenableMessage: '重新启用后，该代理可以继续使用原推广链接并产生佣金。请填写重新启用原因。', restoreMessage: '恢复后该代理可以继续使用推广链接并产生佣金。',
      colUser: '代理用户', colRelation: '层级关系', colCustomers: '客户 / 付费', colPeriodCustomers: '本期新增 / 付费', colRate: '返佣策略', colPaid: '客户实付', colCommission: '佣金收入', colBalance: '佣金余额', colStatus: '状态', colActions: '操作', userInactive: '用户未启用', alreadyAgent: '已经是代理', distributionCustomer: '已绑定代理', affiliateInvitee: '已绑定邀请人', userNormal: '用户正常', unavailable: '当前不可选', loadFailed: '加载代理失败', exportSuccess: '代理数据已导出', exportFailed: '导出代理失败', loadDefaultsFailed: '加载默认返佣设置失败', selectParentError: '请选择所属一级代理', rateRangeError: '返佣比例必须在 0% 到 100% 之间', upgraded: '用户已升级为代理', added: '{level}代理已添加', addFailed: '添加代理失败', loadRateFailed: '加载返佣设置失败', loadEventsFailed: '加载代理变更记录失败', loadAnalyticsFailed: '加载代理经营数据失败', createdEvent: '创建代理', statsGrantedEvent: '开通推广统计权限', statsRevokedEvent: '关闭推广统计权限', recruitmentGrantedEvent: '授予下级招募权限', recruitmentRevokedEvent: '撤销下级招募权限', rateEvent: '返佣比例 {before} → {after}', recruitmentGranted: '已授予下级招募权限', recruitmentRevoked: '已撤销下级招募权限', recruitmentUpdateFailed: '更新招募权限失败', statsGranted: '已开通推广统计权限', statsRevoked: '已关闭推广统计权限', statsUpdateFailed: '更新推广统计权限失败', statusUpdated: '代理状态已更新', statusUpdateFailed: '状态更新失败', rateUpdated: '返佣比例已更新', rateUpdateFailed: '更新返佣比例失败'
    },
  },
}
