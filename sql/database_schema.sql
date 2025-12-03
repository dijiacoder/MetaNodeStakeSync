-- ========================================
-- MetaNodeStake 合约事件同步数据库表结构
-- 数据库版本: MySQL 8.x
-- ========================================

-- 删除已存在的表（开发环境使用，生产环境请注释）
-- DROP TABLE IF EXISTS event_claim;
-- DROP TABLE IF EXISTS event_withdraw;
-- DROP TABLE IF EXISTS event_request_unstake;
-- DROP TABLE IF EXISTS event_deposit;
-- DROP TABLE IF EXISTS event_update_pool;
-- DROP TABLE IF EXISTS event_set_pool_weight;
-- DROP TABLE IF EXISTS event_update_pool_info;
-- DROP TABLE IF EXISTS event_add_pool;
-- DROP TABLE IF EXISTS event_set_metanode_per_block;
-- DROP TABLE IF EXISTS event_set_end_block;
-- DROP TABLE IF EXISTS event_set_start_block;
-- DROP TABLE IF EXISTS event_pause_claim;
-- DROP TABLE IF EXISTS event_pause_withdraw;
-- DROP TABLE IF EXISTS event_set_metanode;
-- DROP TABLE IF EXISTS user_pool_stats;
-- DROP TABLE IF EXISTS pool_info;
-- DROP TABLE IF EXISTS sync_status;

-- ========================================
-- 1. 同步状态表 - 记录区块同步进度
-- ========================================
CREATE TABLE IF NOT EXISTS sync_status (
                                           id INT PRIMARY KEY AUTO_INCREMENT,
                                           contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
    chain_id INT NOT NULL COMMENT '链ID (如 11155111 for Sepolia)',
    last_synced_block BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后同步的区块号',
    last_sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后同步时间',
    sync_error TEXT COMMENT '同步错误信息',
    is_syncing BOOLEAN DEFAULT FALSE COMMENT '是否正在同步',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_contract_chain (contract_address, chain_id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区块链同步状态表';

-- ========================================
-- 2. 资金池信息表 - 存储池的最新状态
-- ========================================
CREATE TABLE IF NOT EXISTS pool_info (
                                         id INT PRIMARY KEY AUTO_INCREMENT,
                                         pool_id INT NOT NULL COMMENT '资金池ID',
                                         contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
    st_token_address VARCHAR(42) NOT NULL COMMENT '质押代币地址 (0x0 表示ETH)',
    pool_weight DECIMAL(30,0) NOT NULL COMMENT '资金池权重',
    last_reward_block BIGINT UNSIGNED NOT NULL COMMENT '最后奖励区块',
    acc_metanode_per_st DECIMAL(65,18) DEFAULT 0 COMMENT '每质押代币累计MetaNode',
    st_token_amount DECIMAL(65,18) DEFAULT 0 COMMENT '质押代币总量',
    min_deposit_amount DECIMAL(65,18) NOT NULL COMMENT '最小质押金额',
    unstake_locked_blocks INT NOT NULL COMMENT '解锁区块数',
    total_pool_weight DECIMAL(30,0) COMMENT '所有池的总权重',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_block BIGINT UNSIGNED COMMENT '创建时的区块号',
    created_tx VARCHAR(66) COMMENT '创建交易哈希',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_pool_contract (pool_id, contract_address),
    INDEX idx_st_token (st_token_address),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资金池信息表';

-- ========================================
-- 3. 用户资金池统计表 - 用户在各个池的实时数据
-- ========================================
CREATE TABLE IF NOT EXISTS user_pool_stats (
                                               id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                               user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    pool_id INT NOT NULL COMMENT '资金池ID',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
    st_amount DECIMAL(65,18) DEFAULT 0 COMMENT '当前质押金额',
    finished_metanode DECIMAL(65,18) DEFAULT 0 COMMENT '已领取的MetaNode',
    pending_metanode DECIMAL(65,18) DEFAULT 0 COMMENT '待领取的MetaNode',
    total_deposited DECIMAL(65,18) DEFAULT 0 COMMENT '累计质押金额',
    total_unstaked DECIMAL(65,18) DEFAULT 0 COMMENT '累计解质押金额',
    total_withdrawn DECIMAL(65,18) DEFAULT 0 COMMENT '累计提现金额',
    total_claimed DECIMAL(65,18) DEFAULT 0 COMMENT '累计领取奖励',
    last_deposit_block BIGINT UNSIGNED COMMENT '最后质押区块',
    last_claim_block BIGINT UNSIGNED COMMENT '最后领取区块',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_pool (user_address, pool_id, contract_address),
    INDEX idx_user (user_address),
    INDEX idx_pool (pool_id),
    INDEX idx_st_amount (st_amount)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户资金池统计表';

-- ========================================
-- 事件表 - 存储所有合约事件
-- ========================================

-- 4. SetMetaNode 事件
CREATE TABLE IF NOT EXISTS event_set_metanode (
                                                  id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                  contract_address VARCHAR(42) NOT NULL,
    metanode_token VARCHAR(42) NOT NULL COMMENT 'MetaNode代币地址',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SetMetaNode事件表';

-- 5. PauseWithdraw / UnpauseWithdraw 事件
CREATE TABLE IF NOT EXISTS event_pause_withdraw (
                                                    id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                    contract_address VARCHAR(42) NOT NULL,
    is_paused BOOLEAN NOT NULL COMMENT '是否暂停 (true=暂停, false=恢复)',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='暂停/恢复提现事件表';

-- 6. PauseClaim / UnpauseClaim 事件
CREATE TABLE IF NOT EXISTS event_pause_claim (
                                                 id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                 contract_address VARCHAR(42) NOT NULL,
    is_paused BOOLEAN NOT NULL COMMENT '是否暂停 (true=暂停, false=恢复)',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='暂停/恢复领取事件表';

-- 7. SetStartBlock 事件
CREATE TABLE IF NOT EXISTS event_set_start_block (
                                                     id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                     contract_address VARCHAR(42) NOT NULL,
    start_block BIGINT UNSIGNED NOT NULL COMMENT '质押开始区块',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置开始区块事件表';

-- 8. SetEndBlock 事件
CREATE TABLE IF NOT EXISTS event_set_end_block (
                                                   id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                   contract_address VARCHAR(42) NOT NULL,
    end_block BIGINT UNSIGNED NOT NULL COMMENT '质押结束区块',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置结束区块事件表';

-- 9. SetMetaNodePerBlock 事件
CREATE TABLE IF NOT EXISTS event_set_metanode_per_block (
                                                            id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                            contract_address VARCHAR(42) NOT NULL,
    metanode_per_block DECIMAL(65,18) NOT NULL COMMENT '每区块MetaNode奖励',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置每区块奖励事件表';

-- 10. AddPool 事件
CREATE TABLE IF NOT EXISTS event_add_pool (
                                              id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                              contract_address VARCHAR(42) NOT NULL,
    pool_id INT NOT NULL COMMENT '资金池ID (根据事件顺序计算)',
    st_token_address VARCHAR(42) NOT NULL COMMENT '质押代币地址',
    pool_weight DECIMAL(30,0) NOT NULL COMMENT '资金池权重',
    last_reward_block BIGINT UNSIGNED NOT NULL COMMENT '最后奖励区块',
    min_deposit_amount DECIMAL(65,18) NOT NULL COMMENT '最小质押金额',
    unstake_locked_blocks INT NOT NULL COMMENT '解锁区块数',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_pool (pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='添加资金池事件表';

-- 11. UpdatePoolInfo 事件
CREATE TABLE IF NOT EXISTS event_update_pool_info (
                                                      id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                      contract_address VARCHAR(42) NOT NULL,
    pool_id INT NOT NULL COMMENT '资金池ID',
    min_deposit_amount DECIMAL(65,18) NOT NULL COMMENT '最小质押金额',
    unstake_locked_blocks INT NOT NULL COMMENT '解锁区块数',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_pool (pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='更新资金池信息事件表';

-- 12. SetPoolWeight 事件
CREATE TABLE IF NOT EXISTS event_set_pool_weight (
                                                     id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                     contract_address VARCHAR(42) NOT NULL,
    pool_id INT NOT NULL COMMENT '资金池ID',
    pool_weight DECIMAL(30,0) NOT NULL COMMENT '新的资金池权重',
    total_pool_weight DECIMAL(30,0) NOT NULL COMMENT '所有池的总权重',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_pool (pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置资金池权重事件表';

-- 13. UpdatePool 事件
CREATE TABLE IF NOT EXISTS event_update_pool (
                                                 id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                 contract_address VARCHAR(42) NOT NULL,
    pool_id INT NOT NULL COMMENT '资金池ID',
    last_reward_block BIGINT UNSIGNED NOT NULL COMMENT '最后奖励区块',
    total_metanode DECIMAL(65,18) NOT NULL COMMENT '本次更新的总MetaNode奖励',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_pool (pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='更新资金池事件表';

-- 14. Deposit 事件 (最重要的用户操作事件之一)
CREATE TABLE IF NOT EXISTS event_deposit (
                                             id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                             contract_address VARCHAR(42) NOT NULL,
    user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    pool_id INT NOT NULL COMMENT '资金池ID',
    amount DECIMAL(65,18) NOT NULL COMMENT '质押金额',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_user (user_address),
    INDEX idx_pool (pool_id),
    INDEX idx_user_pool (user_address, pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押事件表';

-- 15. RequestUnstake 事件
CREATE TABLE IF NOT EXISTS event_request_unstake (
                                                     id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                     contract_address VARCHAR(42) NOT NULL,
    user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    pool_id INT NOT NULL COMMENT '资金池ID',
    amount DECIMAL(65,18) NOT NULL COMMENT '解质押金额',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_user (user_address),
    INDEX idx_pool (pool_id),
    INDEX idx_user_pool (user_address, pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='申请解质押事件表';

-- 16. Withdraw 事件
CREATE TABLE IF NOT EXISTS event_withdraw (
                                              id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                              contract_address VARCHAR(42) NOT NULL,
    user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    pool_id INT NOT NULL COMMENT '资金池ID',
    amount DECIMAL(65,18) NOT NULL COMMENT '提现金额',
    withdraw_block_number BIGINT UNSIGNED NOT NULL COMMENT '提现时的区块号',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_user (user_address),
    INDEX idx_pool (pool_id),
    INDEX idx_user_pool (user_address, pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现事件表';

-- 17. Claim 事件 (最重要的用户操作事件之一)
CREATE TABLE IF NOT EXISTS event_claim (
                                           id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                           contract_address VARCHAR(42) NOT NULL,
    user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    pool_id INT NOT NULL COMMENT '资金池ID',
    metanode_reward DECIMAL(65,18) NOT NULL COMMENT '领取的MetaNode奖励',
    block_number BIGINT UNSIGNED NOT NULL,
    block_timestamp BIGINT UNSIGNED NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tx_log (transaction_hash, log_index),
    INDEX idx_block (block_number),
    INDEX idx_user (user_address),
    INDEX idx_pool (pool_id),
    INDEX idx_user_pool (user_address, pool_id),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='领取奖励事件表';

-- ========================================
-- 18. 用户解质押请求表 - 记录待提现的解质押记录
-- ========================================
CREATE TABLE IF NOT EXISTS user_unstake_requests (
                                                     id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                                     user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    pool_id INT NOT NULL COMMENT '资金池ID',
    contract_address VARCHAR(42) NOT NULL,
    amount DECIMAL(65,18) NOT NULL COMMENT '解质押金额',
    unlock_block BIGINT UNSIGNED NOT NULL COMMENT '解锁区块号',
    request_block BIGINT UNSIGNED NOT NULL COMMENT '申请时的区块号',
    request_tx VARCHAR(66) NOT NULL COMMENT '申请交易哈希',
    is_withdrawn BOOLEAN DEFAULT FALSE COMMENT '是否已提现',
    withdrawn_block BIGINT UNSIGNED COMMENT '提现区块号',
    withdrawn_tx VARCHAR(66) COMMENT '提现交易哈希',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_pool (user_address, pool_id),
    INDEX idx_unlock_block (unlock_block),
    INDEX idx_withdrawn (is_withdrawn),
    INDEX idx_contract (contract_address)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户解质押请求表';

-- ========================================
-- 19. 统计视图 - 便于查询
-- ========================================

-- 用户总览统计视图
CREATE OR REPLACE VIEW v_user_stats AS
SELECT
    u.user_address,
    u.contract_address,
    COUNT(DISTINCT u.pool_id) as pool_count,
    SUM(u.st_amount) as total_staked,
    SUM(u.total_deposited) as total_deposited,
    SUM(u.total_claimed) as total_claimed,
    SUM(u.pending_metanode) as total_pending_reward,
    MAX(u.updated_at) as last_activity
FROM user_pool_stats u
GROUP BY u.user_address, u.contract_address;

-- 资金池统计视图
CREATE OR REPLACE VIEW v_pool_stats AS
SELECT
    p.pool_id,
    p.contract_address,
    p.st_token_address,
    p.pool_weight,
    p.st_token_amount as total_staked,
    COUNT(DISTINCT u.user_address) as user_count,
    (SELECT COUNT(*) FROM event_deposit d WHERE d.pool_id = p.pool_id AND d.contract_address = p.contract_address) as deposit_count,
    (SELECT SUM(amount) FROM event_deposit d WHERE d.pool_id = p.pool_id AND d.contract_address = p.contract_address) as total_deposits,
    (SELECT SUM(metanode_reward) FROM event_claim c WHERE c.pool_id = p.pool_id AND c.contract_address = p.contract_address) as total_claimed_rewards
FROM pool_info p
         LEFT JOIN user_pool_stats u ON p.pool_id = u.pool_id AND p.contract_address = u.contract_address
GROUP BY p.pool_id, p.contract_address, p.st_token_address, p.pool_weight, p.st_token_amount;

-- ========================================
-- 初始化数据
-- ========================================

-- 插入默认同步状态记录（请根据实际情况修改）
-- INSERT INTO sync_status (contract_address, chain_id, last_synced_block)
-- VALUES ('0xYourContractAddress', 11155111, 0)
-- ON DUPLICATE KEY UPDATE last_synced_block = last_synced_block;
