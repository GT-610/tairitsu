export const setupWizardDatabaseStepCopy = {
  description: '当前仅支持 SQLite 单实例部署。数据库文件将由程序使用默认路径管理。',
  supportAlert: 'MySQL 与 PostgreSQL 相关抽象暂时保留，但当前不在支持范围内。',
  databaseTypeHelperText: '当前固定为 SQLite',
  databasePathHelperText: '留空则使用默认值 data/tairitsu.db',
}
