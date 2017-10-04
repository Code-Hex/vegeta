import * as request from 'superagent';

class Render {
    private _token: string = ""
    constructor() {
        let e = <HTMLInputElement>document.getElementById('api-token')
        this._token = e.value
    }
    
    public get token(): string {
        return this._token
    }

    public GenerateGraph(id: Number): void {
        this.getdata(id)
    }

    private getdata(id: Number): void {
        request.post('/mypage/api/data')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({ tag_id: id })
        .end(function(err, res){
            if (err || !res.ok) {
                alert('http error: ' + err);
            } else {
                let json = res.body
                if (json.is_success) {
                    console.log(json.data)
                    //window.location.reload(true)
                } else {
                    alert(`データの取得に失敗しました: ${ json.reason }`)
                    window.location.reload(true)
                }
            }
        })
    }
}

var render = new Render()

var actions = <NodeListOf<HTMLElement>>document.querySelectorAll('.action')
for (let i = 0; i < actions.length; ++i) {
    actions[i].addEventListener('click', function(e) {
        e.preventDefault()
        let name = this.getAttribute("name")
        if (name == null) {
            console.log("id is null")
            return
        }
        let id = Number(name)
        render.GenerateGraph(id)
        console.log(id)
    })
}

/*
var chart = c3.generate({
    data: {
        x: 'date',
        xFormat: '%Y/%m/%d', // 'xFormat' can be used as custom format of 'x'
        json: [
            {date: '2015/10/30', upload: 0.200, download: 0.200, total: 0.400},
            {date: '2015/10/31', upload: 0.100, download: 0.300, total: 0.400},
            {date: '2015/11/01', upload: 0.300, download: 0.200, total: 0.500},
            {date: '2015/11/03', upload: 0.400, download: 0.100, total: 0.500},
            {date: '2015/11/04', upload: 0.300, download: 0.200, total: 0.500},
            {date: '2015/11/05', upload: 0.400, download: 0.100, total: 0.500},
            {date: '2015/11/06', upload: 0.300, download: 0.200, total: 0.500},
            {date: '2015/11/07', upload: 0.400, download: 0.200, total: 0.600},
            {date: '2015/11/09', upload: 0.300, download: 0.100, total: 0.400},
            {date: '2015/11/12', upload: 0.400, download: 0.100, total: 0.500},
            {date: '2015/11/14', upload: 0.300, download: 0.200, total: 0.500},
            //{date: '2015/11/15', upload: 400, download: 0, total: 400}
        ],
        keys: {
            x: 'date', // it's possible to specify 'x' when category axis
            value: ['upload', 'download', 'total']
        }
    },
    axis: {
        x: {
            type: 'timeseries',
            tick: {
                format: '%Y-%m-%d'
            }
        }
    }
})
*/