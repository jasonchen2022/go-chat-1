CREATE TABLE `article`
(
    `id`          int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章ID',
    `user_id`     int(11) unsigned NOT NULL COMMENT '用户ID',
    `class_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '分类ID',
    `tags_id`     varchar(50)  NOT NULL DEFAULT '' COMMENT '笔记关联标签',
    `title`       varchar(80)  NOT NULL DEFAULT '' COMMENT '文章标题',
    `abstract`    varchar(200) NOT NULL DEFAULT '' COMMENT '文章摘要',
    `image`       varchar(255) NOT NULL DEFAULT '' COMMENT '文章首图',
    `is_asterisk` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否星标文章[0:否;1:是;]',
    `status`      tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '笔记状态[1:正常;2:已删除;]',
    `created_at`  datetime     NOT NULL COMMENT '创建时间',
    `updated_at`  datetime     NOT NULL COMMENT '更新时间',
    `deleted_at`  datetime              DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY           `idx_userid_classid_title` (`user_id`,`class_id`,`title`)
) ENGINE=InnoDB AUTO_INCREMENT=399 DEFAULT CHARSET=utf8 COMMENT='用户文章表';;

CREATE TABLE `article_annex`
(
    `id`            int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文件ID',
    `user_id`       int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上传文件的用户ID',
    `article_id`    int(11) unsigned NOT NULL DEFAULT '1' COMMENT '笔记ID',
    `drive`         tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '文件驱动[1:local;2:cos;]',
    `suffix`        varchar(10)  NOT NULL DEFAULT '' COMMENT '文件后缀名',
    `size`          bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
    `path`          varchar(500) NOT NULL DEFAULT '' COMMENT '文件地址（相对地址）',
    `original_name` varchar(100) NOT NULL DEFAULT '' COMMENT '原文件名',
    `status`        tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '附件状态[1:正常;2:已删除;]',
    `created_at`    datetime     NOT NULL COMMENT '创建时间',
    `updated_at`    datetime     NOT NULL COMMENT '更新时间',
    `deleted_at`    datetime              DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY             `idx_userid_articleid` (`user_id`,`article_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=404 DEFAULT CHARSET=utf8 COMMENT='文章附件信息表';;

CREATE TABLE `article_class`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章分类ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `class_name` varchar(20) NOT NULL DEFAULT '' COMMENT '分类名',
    `sort`       tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '排序',
    `is_default` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '默认分类[0:否;1:是；]',
    `created_at` datetime    NOT NULL COMMENT '创建时间',
    `updated_at` datetime    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id_class_name` (`user_id`,`class_name`) USING BTREE,
    KEY          `uk_user_id_sort` (`user_id`,`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=3352 DEFAULT CHARSET=utf8 COMMENT='文章分类表';;

CREATE TABLE `article_detail`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章详情ID',
    `article_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '文章ID',
    `md_content` longtext CHARACTER SET utf8mb4 NOT NULL COMMENT 'Markdown 内容',
    `content`    longtext CHARACTER SET utf8mb4 NOT NULL COMMENT 'Markdown 解析HTML内容',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_article_id` (`article_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=396 DEFAULT CHARSET=utf8 COMMENT='文章详情表';;

CREATE TABLE `article_tag`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章分类ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `tag_name`   varchar(20) NOT NULL DEFAULT '' COMMENT '标签名',
    `sort`       tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '排序',
    `created_at` datetime    NOT NULL COMMENT '创建时间',
    `updated_at` datetime    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY          `idx_userid` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=69 DEFAULT CHARSET=utf8 COMMENT='文章标签表';;

CREATE TABLE `contact`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '关系ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
    `friend_id`  int(11) unsigned NOT NULL DEFAULT '0' COMMENT '好友id',
    `remark`     varchar(20) NOT NULL DEFAULT '' COMMENT '好友的备注',
    `status`     tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '好友状态 [0:否;1:是]',
    `created_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY          `idx_user_id_friend_id` (`user_id`,`friend_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=269742 DEFAULT CHARSET=utf8 COMMENT='用户好友关系表';;

CREATE TABLE `contact_apply`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '申请ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '申请人ID',
    `friend_id`  int(11) unsigned NOT NULL DEFAULT '0' COMMENT '被申请人',
    `remark`     varchar(50) NOT NULL DEFAULT '' COMMENT '申请备注',
    `created_at` datetime    NOT NULL COMMENT '申请时间',
    PRIMARY KEY (`id`),
    KEY          `idx_user_id` (`user_id`) USING BTREE,
    KEY          `idx_friend_id` (`friend_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=257 DEFAULT CHARSET=utf8 COMMENT='用户添加好友申请表';;

CREATE TABLE `emoticon`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '表情分组ID',
    `name`       varchar(50)  NOT NULL DEFAULT '' COMMENT '分组名称',
    `icon`       varchar(255) NOT NULL DEFAULT '' COMMENT '分组图标',
    `status`     tinyint(4) NOT NULL DEFAULT '0' COMMENT '分组状态[-1:已删除;0:正常;1:已禁用;]',
    `created_at` datetime     NOT NULL COMMENT '创建时间',
    `updated_at` datetime     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8 COMMENT='表情包分组';;

CREATE TABLE `emoticon_item`
(
    `id`          int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '表情包详情ID',
    `emoticon_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '表情分组ID',
    `user_id`     int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID（0：代码系统表情包）',
    `describe`    varchar(20)  NOT NULL DEFAULT '' COMMENT '表情描述',
    `url`         varchar(255) NOT NULL DEFAULT '' COMMENT '图片链接',
    `file_suffix` varchar(10)  NOT NULL DEFAULT '' COMMENT '文件后缀名',
    `file_size`   bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '文件大小（单位字节）',
    `created_at`  datetime     NOT NULL COMMENT '创建时间',
    `updated_at`  datetime     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=141 DEFAULT CHARSET=utf8 COMMENT='表情包详情表';;

CREATE TABLE `group`
(
    `id`           int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '群ID',
    `type`         tinyint(4) unsigned NOT NULL DEFAULT '1' COMMENT '群类型[1:普通群;2:企业群;]',
    `creator_id`   int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建者ID(群主ID)',
    `group_name`   varchar(30) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT '群名称',
    `profile`      varchar(100)                      NOT NULL DEFAULT '' COMMENT '群介绍',
    `is_dismiss`   tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否已解散[0:否;1:是;]',
    `avatar`       varchar(255)                      NOT NULL DEFAULT '' COMMENT '群头像',
    `max_num`      smallint(5) unsigned NOT NULL DEFAULT '200' COMMENT '最大群成员数量',
    `is_overt`     tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否公开可见[0:否;1:是;]',
    `is_mute`      tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否全员禁言 [0:否;1:是;]，提示:不包含群主或管理员',
    `created_at`   datetime                          NOT NULL COMMENT '创建时间',
    `updated_at`   datetime                          NOT NULL COMMENT '更新时间',
    `dismissed_at` datetime                                   DEFAULT NULL COMMENT '解散时间',
    PRIMARY KEY (`id`),
    KEY            `idx_created_at` (`created_at`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=421 DEFAULT CHARSET=utf8 COMMENT='用户聊天群';;

CREATE TABLE `group_apply`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `group_id`   int(11) unsigned NOT NULL DEFAULT '0' COMMENT '群组ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `remark`     varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT '备注信息',
    `created_at` datetime                           NOT NULL COMMENT '创建时间',
    `updated_at` datetime                           NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY          `idx_group_id_user_id` (`group_id`,`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2439 DEFAULT CHARSET=utf8 COMMENT='群聊成员';;

CREATE TABLE `group_member`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `group_id`   int(11) unsigned NOT NULL DEFAULT '0' COMMENT '群组ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `leader`     tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '成员属性[0:普通成员;1:管理员;2:群主;]',
    `user_card`  varchar(20) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT '群名片',
    `is_quit`    tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否退群[0:否;1:是;]',
    `is_mute`    tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否禁言[0:否;1:是;]',
    `created_at` datetime                          NOT NULL COMMENT '创建时间',
    `updated_at` datetime                          NOT NULL COMMENT '更新时间',
    `deleted_at` datetime                                   DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_group_id_user_id` (`group_id`,`user_id`) USING BTREE,
    KEY          `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2909 DEFAULT CHARSET=utf8 COMMENT='群聊成员';;

CREATE TABLE `group_notice`
(
    `id`            int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '公告ID',
    `group_id`      int(11) unsigned NOT NULL DEFAULT '0' COMMENT '群组ID',
    `creator_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建者用户ID',
    `title`         varchar(50) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT '公告标题',
    `content`       text CHARACTER SET utf8mb4 NOT NULL COMMENT '公告内容',
    `confirm_users` json                                       DEFAULT NULL COMMENT '已确认成员',
    `is_delete`     tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否删除[0:否;1:是;]',
    `is_top`        tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否置顶[0:否;1:是;]',
    `is_confirm`    tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否需群成员确认公告[0:否;1:是;]',
    `created_at`    datetime                          NOT NULL COMMENT '创建时间',
    `updated_at`    datetime                          NOT NULL COMMENT '更新时间',
    `deleted_at`    datetime                                   DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY             `idx_group` (`group_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=66 DEFAULT CHARSET=utf8 COMMENT='群组公告表';;

CREATE TABLE `organize`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
    `department` varchar(100) NOT NULL DEFAULT '' COMMENT '部门ID',
    `position`   varchar(100) NOT NULL DEFAULT '' COMMENT '岗位ID',
    `created_at` datetime     NOT NULL COMMENT '创建时间',
    `updated_at` datetime     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8 COMMENT='组织表';;

CREATE TABLE `organize_dept`
(
    `dept_id`    int(11) NOT NULL AUTO_INCREMENT COMMENT '部门id',
    `parent_id`  int(11) NOT NULL DEFAULT '0' COMMENT '父部门id',
    `ancestors`  varchar(50) NOT NULL DEFAULT '' COMMENT '祖级列表',
    `dept_name`  varchar(30) NOT NULL DEFAULT '' COMMENT '部门名称',
    `order_num`  int(4) NOT NULL DEFAULT '0' COMMENT '显示顺序',
    `leader`     varchar(20) NOT NULL COMMENT '负责人',
    `phone`      varchar(11) NOT NULL COMMENT '联系电话',
    `email`      varchar(50) NOT NULL COMMENT '邮箱',
    `status`     tinyint(4) NOT NULL DEFAULT '1' COMMENT '部门状态[1:正常;2:停用]',
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '1' COMMENT '是否删除[1:否;2:是]',
    `created_at` datetime    NOT NULL COMMENT '创建时间',
    `updated_at` datetime    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`dept_id`)
) ENGINE=InnoDB AUTO_INCREMENT=111 DEFAULT CHARSET=utf8 COMMENT='部门表';;

CREATE TABLE `organize_position`
(
    `position_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '岗位ID',
    `post_code`   varchar(30)  NOT NULL COMMENT '岗位编码',
    `post_name`   varchar(50)  NOT NULL COMMENT '岗位名称',
    `sort`        int(4) unsigned NOT NULL DEFAULT '0' COMMENT '显示顺序',
    `status`      tinyint(4) unsigned NOT NULL DEFAULT '1' COMMENT '状态[1:正常;2:停用;]',
    `remark`      varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
    `created_at`  datetime     NOT NULL COMMENT '创建时间',
    `updated_at`  datetime     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`position_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 COMMENT='岗位信息表';;

CREATE TABLE `robot`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '机器人ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '关联用户ID',
    `robot_name` varchar(20)  NOT NULL DEFAULT '' COMMENT '机器人名称',
    `describe`   varchar(255) NOT NULL DEFAULT '' COMMENT '描述信息',
    `logo`       varchar(255) NOT NULL DEFAULT '' COMMENT '机器人logo',
    `is_talk`    tinyint(4) NOT NULL DEFAULT '0' COMMENT '可发送消息[0:否;1:是;]',
    `status`     tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '状态[-1:已删除;0:正常;1:已禁用;]',
    `type`       tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '机器人类型',
    `created_at` datetime     NOT NULL COMMENT '创建时间',
    `updated_at` datetime     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_type` (`type`) USING HASH,
    UNIQUE KEY `uk_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COMMENT='聊天机器人表';;


CREATE TABLE `split_upload`
(
    `id`            int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '临时文件ID',
    `type`          tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '文件属性[1:合并文件;2:拆分文件]',
    `drive`         tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '驱动类型[1:local;2:cos;]',
    `upload_id`     varchar(100) NOT NULL DEFAULT '' COMMENT '临时文件hash名',
    `user_id`       int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上传的用户ID',
    `original_name` varchar(100) NOT NULL DEFAULT '' COMMENT '原文件名',
    `split_index`   int(11) unsigned NOT NULL DEFAULT '0' COMMENT '当前索引块',
    `split_num`     int(11) unsigned NOT NULL DEFAULT '0' COMMENT '总上传索引块',
    `path`          varchar(255) NOT NULL DEFAULT '' COMMENT '临时保存路径',
    `file_ext`      varchar(10)  NOT NULL DEFAULT '' COMMENT '文件后缀名',
    `file_size`     int(11) unsigned NOT NULL COMMENT '文件大小',
    `is_delete`     tinyint(2) unsigned NOT NULL DEFAULT '0' COMMENT '文件是否删除[0:否;1:是;] ',
    `attr`          json         NOT NULL COMMENT '额外参数json',
    `created_at`    datetime     NOT NULL COMMENT '更新时间',
    `updated_at`    datetime     NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY             `idx_user_id_hash_name` (`user_id`,`upload_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3509 DEFAULT CHARSET=utf8 COMMENT='文件拆分数据表';;

CREATE TABLE `talk_records`
(
    `id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '聊天记录ID',
    `talk_type`   tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '对话类型[1:私信;2:群聊;]',
    `msg_type`    tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '消息类型[1:文本消息;2:文件消息;3:会话消息;4:代码消息;5:投票消息;6:群公告;7:好友申请;8:登录通知;9:入群消息/退群消息;]',
    `user_id`     int(11) unsigned NOT NULL DEFAULT '0' COMMENT '发送者ID（0:代表系统消息 >0: 用户ID）',
    `receiver_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '接收者ID（用户ID 或 群ID）',
    `is_revoke`   tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否撤回消息[0:否;1:是;]',
    `is_mark`     tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否重要消息[0:否;1:是;]',
    `is_read`     tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否已读[0:否;1:是;]',
    `quote_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '引用消息ID',
    `warn_users`  varchar(200) NOT NULL DEFAULT '' COMMENT '@好友 、 多个用英文逗号 “,” 拼接 (0:代表所有人)',
    `content`     text CHARACTER SET utf8mb4 COMMENT '文本消息 {@nickname@}',
    `created_at`  datetime     NOT NULL COMMENT '创建时间',
    `updated_at`  datetime     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY           `idx_user_id_receiver_id` (`user_id`,`receiver_id`)
) ENGINE=InnoDB AUTO_INCREMENT=211754 DEFAULT CHARSET=utf8 COMMENT='用户聊天记录表';;

CREATE TABLE `talk_records_code`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '代码块ID',
    `record_id`  bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `lang`       varchar(20) NOT NULL DEFAULT '' COMMENT '语言类型(如：php,java,python)',
    `code`       text CHARACTER SET utf8mb4 NOT NULL COMMENT '代码内容',
    `created_at` datetime    NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY          `idx_record_id` (`record_id`)
) ENGINE=InnoDB AUTO_INCREMENT=681 DEFAULT CHARSET=utf8 COMMENT='用户聊天记录（代码块消息）';;

CREATE TABLE `talk_records_delete`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
    `record_id`  int(11) unsigned NOT NULL DEFAULT '0' COMMENT '聊天记录ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `created_at` datetime NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_record_user_id` (`record_id`,`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=368 DEFAULT CHARSET=utf8 COMMENT='聊天记录删除记录表';;

CREATE TABLE `talk_records_file`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '文件ID',
    `record_id`     bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `user_id`       int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上传文件的用户ID',
    `source`        tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '文件来源[1:用户上传;2:表情包;]',
    `type`          tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '文件类型[1:图片;2:音频文件;3:视频文件;4:其它文件;]',
    `drive`         tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '驱动类型[1:local;2:cos;]',
    `original_name` varchar(100) NOT NULL DEFAULT '' COMMENT '原文件名',
    `suffix`        varchar(10)  NOT NULL DEFAULT '' COMMENT '文件后缀',
    `size`          bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
    `path`          varchar(300) NOT NULL DEFAULT '' COMMENT '文件地址(相对地址)',
    `url`           varchar(255) NOT NULL DEFAULT '' COMMENT '网络地址(公开文件地址)',
    `created_at`    datetime     NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_record_id` (`record_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2545 DEFAULT CHARSET=utf8 COMMENT='用户聊天记录（文件消息）';;

CREATE TABLE `talk_records_forward`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '合并转发ID',
    `record_id`  bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '转发用户ID',
    `records_id` varchar(255) NOT NULL DEFAULT '' COMMENT '转发的聊天记录ID （多个用 , 拼接），最多只能30条记录信息',
    `text`       json         NOT NULL COMMENT '记录快照（避免后端再次查询数据）',
    `created_at` datetime     NOT NULL COMMENT '转发时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_records_id` (`record_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=178 DEFAULT CHARSET=utf8 COMMENT='用户聊天记录（会话记录转发消息）';;

CREATE TABLE `talk_records_invite`
(
    `id`              int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '入群或退群通知ID',
    `record_id`       bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `type`            tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '通知类型 （1:入群通知 2:自动退群 3:管理员踢群）',
    `operate_user_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '操作人的用户ID（邀请人OR管理员ID）',
    `user_ids`        varchar(255) NOT NULL COMMENT '用户ID，多个用 , 分割',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_recordid` (`record_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=473 DEFAULT CHARSET=utf8 COMMENT='用户聊天记录（入群/退群消息）';;

CREATE TABLE `talk_records_location`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `record_id`  int(11) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `longitude`  decimal(11, 6) NOT NULL DEFAULT '0.000000' COMMENT '经度',
    `latitude`   decimal(11, 6) NOT NULL DEFAULT '0.000000' COMMENT '纬度',
    `created_at` datetime       NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_record_id` (`record_id`) USING BTREE,
    KEY          `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8 COMMENT='聊天对话记录（位置消息）';;

CREATE TABLE `talk_records_login`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '登录ID',
    `record_id`  int(11) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `ip`         varchar(20)  NOT NULL DEFAULT '' COMMENT 'IP地址',
    `platform`   varchar(20)  NOT NULL DEFAULT '' COMMENT '登录平台[h5,ios,windows,mac,web]',
    `agent`      varchar(500) NOT NULL DEFAULT '' COMMENT '设备信息',
    `address`    varchar(100) NOT NULL DEFAULT '' COMMENT 'IP所在地',
    `reason`     varchar(100) NOT NULL DEFAULT '' COMMENT '异常提示',
    `created_at` datetime     NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_record_id` (`record_id`) USING BTREE,
    KEY          `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=28562 DEFAULT CHARSET=utf8 COMMENT='聊天对话记录（登录日志）';;

CREATE TABLE `talk_records_vote`
(
    `id`            int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '投票ID',
    `record_id`     int(11) unsigned NOT NULL DEFAULT '0' COMMENT '消息记录ID',
    `user_id`       int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `title`         varchar(50) NOT NULL DEFAULT '' COMMENT '投票标题',
    `answer_mode`   tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '答题模式[0:单选;1:多选;]',
    `answer_option` json        NOT NULL COMMENT '答题选项',
    `answer_num`    smallint(6) unsigned NOT NULL DEFAULT '0' COMMENT '应答人数',
    `answered_num`  smallint(6) unsigned NOT NULL DEFAULT '0' COMMENT '已答人数',
    `is_anonymous`  tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '匿名投票[0:否;1:是;]',
    `status`        tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '投票状态[0:投票中;1:已完成;]',
    `created_at`    datetime    NOT NULL COMMENT '创建时间',
    `updated_at`    datetime    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_record_id` (`record_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=117 DEFAULT CHARSET=utf8 COMMENT='聊天对话记录（投票消息表）';;

CREATE TABLE `talk_records_vote_answer`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '答题ID',
    `vote_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '投票ID',
    `user_id`    int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `option`     char(1)  NOT NULL DEFAULT '' COMMENT '投票选项[A、B、C 、D、E、F]',
    `created_at` datetime NOT NULL COMMENT '答题时间',
    PRIMARY KEY (`id`),
    KEY          `idx_vote_id_user_id` (`vote_id`,`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=172 DEFAULT CHARSET=utf8 COMMENT='聊天对话记录（投票消息统计表）';;

CREATE TABLE `talk_session`
(
    `id`          int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '聊天列表ID',
    `talk_type`   tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '聊天类型[1:私信;2:群聊;]',
    `user_id`     int(11) NOT NULL DEFAULT '0' COMMENT '用户ID',
    `receiver_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '接收者ID（用户ID 或 群ID）',
    `is_top`      tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否置顶[0:否;1:是;]',
    `is_disturb`  tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '消息免打扰[0:否;1:是;]',
    `is_delete`   tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否删除[0:否;1:是;]',
    `is_robot`    tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否机器人[0:否;1:是;]',
    `created_at`  datetime NOT NULL COMMENT '创建时间',
    `updated_at`  datetime NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_id_receiver_id_talk_type` (`user_id`,`receiver_id`,`talk_type`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1576 DEFAULT CHARSET=utf8 COMMENT='会话列表';;

CREATE TABLE `users`
(
    `id`         int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `mobile`     varchar(11)                       NOT NULL DEFAULT '' COMMENT '手机号',
    `nickname`   varchar(20) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT '用户昵称',
    `avatar`     varchar(255)                      NOT NULL DEFAULT '' COMMENT '用户头像地址',
    `gender`     tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '用户性别[0:未知;1:男 ;2:女;]',
    `password`   varchar(255)                      NOT NULL COMMENT '用户密码',
    `motto`      varchar(100)                      NOT NULL DEFAULT '' COMMENT '用户座右铭',
    `email`      varchar(30)                       NOT NULL DEFAULT '' COMMENT '用户邮箱',
    `is_robot`   tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否机器人[0:否;1:是;]',
    `created_at` datetime                          NOT NULL COMMENT '注册时间',
    `updated_at` datetime                          NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `idx_mobile` (`mobile`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=4361 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户表';;

CREATE TABLE `users_emoticon`
(
    `id`           int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '表情包收藏ID',
    `user_id`      int(11) unsigned NOT NULL COMMENT '用户ID',
    `emoticon_ids` varchar(255) NOT NULL DEFAULT '' COMMENT '表情包ID',
    `created_at`   datetime     NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 COMMENT='用户收藏表情包';;

INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (1, '10046798935', '登录助手', '', 0, '$2y$10$4XW5vq07jVoRUJUfGHYDUeHWcPjFDlC7bVwHe9wplv5Ors2dZilau', '', '', 1,
        '2022-07-12 20:24:01', '2022-07-12 20:24:01');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (2, '18798272050', 'test0', '', 0, '$2y$10$Om135sncVgfj26ISd.TXGuOOboJLC3gdv1cUtY1Rojc20NUCUFrzC', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (3, '18798272051', 'test1', '', 0, '$2y$10$P3/4ya2lJ.nFf48yv.OuxO58rIsXM28Oa0fClHzOc0XOOFtg9IXKW', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (4, '18798272052', 'test2', '', 0, '$2y$10$EC4rqwwhEUKs5eNWB4ciEu20WzkoT7wzK4VcKBgp/a38ahNYSSVEa', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (5, '18798272053', 'test3', '', 0, '$2y$10$R1vRWkARgL8MWDQnewakOODAeOlJ6JLMQ/6jyFLM/cykC5ySgaW9q', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (6, '18798272054', 'test4', '', 0, '$2y$10$P/P5JEiUKg5TS0UzCyr4NuX8NzL9qgx.xdlInCD0g.uoVoKk8ncWm', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (7, '18798272055', 'test5', '', 0, '$2y$10$9y9QuZHDYKEtK85Vz9f7A.CqHS3bGUrppOMuA8X5z1CmV0VrEwAMi', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (8, '18798272056', 'test6', '', 0, '$2y$10$LP7tDHXi.SK0m/cTdoNX1O8hYMp08OdcNfDPoB90ylOkJJEZSLo7O', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (9, '18798272057', 'test7', '', 0, '$2y$10$1AQ0JpD70ro6Khw45DxX4ucAD7OpdkyNI7VpeA0ag.gkJyYSSf4w2', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;
INSERT INTO `users`(`id`, `mobile`, `nickname`, `avatar`, `gender`, `password`, `motto`, `email`, `is_robot`,
                    `created_at`, `updated_at`)
VALUES (10, '18798272058', 'test8', '', 0, '$2y$10$zt7NlMaV8Z1UvzvU8B3AD.q9e/5nKah1Lttpz6BZRy7KL.DA.c3J2', '', '', 0,
        '2022-07-12 20:24:49', '2022-07-12 20:24:49');;