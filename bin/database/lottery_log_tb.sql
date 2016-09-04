drop database if exists `lottery_log`;
create database if not exists `lottery_log`;

use `lottery_log`;

create table `lottery_log_tb`(
	cardID varchar(128) not null,
	userID varchar(128) not null,
	costScore int unsigned default 0,
	rewardScore int unsigned default 0,
	rewardRate int unsigned default 0,
	opera_time datetime not null
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
