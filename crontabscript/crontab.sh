#!/bin/bash
port=$(pidof rsync)
Date=$(date +%F" "%T)
IP=$(ifconfig ens192 |grep "inet " |awk '{print $2}')
MASTER="金蝶测试环境 服务器"
COMMENT="rsync已经传输完成"
if [[ -z $port ]];then
   /data/webhook/wechatwebhook  $Date $MASTER $IP $COMMENT
   crontab -r
   #echo "yes"
else
   echo "test sucesss"
fi
