import * as request from 'superagent';
import * as c3 from 'c3';

var chart = c3.generate({
    data: {
        x: 'x',
        xFormat: '%Y/%m/%d %H:%M:%S', // 'xFormat' can be used as custom format of 'x'
        columns: [
            ['x', '2015/10/30', '2015/10/31', '2015/11/01', '2015/11/02', '2015/11/03', '2015/11/04'],
            ['data1', 30, 200, 100, 400, 150, 250],
            ['data2', 130, 340, 200, 500, 250, 350]
        ]
    },
    axis: {
        x: {
            type: 'timeseries',
            tick: {
                rotate: 45,
                format: '%Y/%m/%d'
            }
        }
    }
})
