/**
 * Created by meicai on 16-1-6.
 */

var BackupTypes = {
    0: '全量',
    1: '增量',
    2: '压缩'
};

var CronJobStatus = {
    0: '运行中',
    1: '休息中',
    2: '运行出错'
};

var formatter = {
    getDatetime: function (datetimestr) {
        var fixLength = function (value) {
            return value > 9 ? value.toString() : '0' + value.toString()
        };
        var d = new Date(datetimestr);
        return [d.getFullYear(), fixLength(d.getMonth() + 1), fixLength(d.getDate())].join('-') + ' ' + [fixLength(d.getHours()), fixLength(d.getMinutes()), fixLength(d.getSeconds())].join(':')
    },

    getDisplayForType: function(type){
        if (BackupTypes.hasOwnProperty(type)){
            return BackupTypes[type]
        }else{
            return '未知';
        }
    },

    getUptime: function(since){
        var start = new Date(since);
        var now = new Date();
        var millSeconds = now - start;
        return Math.floor(millSeconds / 1000 / 60 / 60) + '小时，' + Math.floor(millSeconds / 1000 / 60) % 60 + '分钟，' + Math.floor(millSeconds / 1000) % 60 + '秒';
    }
};