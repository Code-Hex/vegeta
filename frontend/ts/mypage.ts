import * as request from 'superagent';
import * as c3 from 'c3';

class Render {
    private _token: string = ""
    private _datePtn = /^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}:[0-9]{2}:[0-9]{2})\+[0-9]{2}:[0-9]{2}$/
    constructor() {
        let e = <HTMLInputElement>document.getElementById('api-token')
        this._token = e.value
    }
    
    public get token(): string {
        return this._token
    }

    public GenerateGraph(id: Number): void {
        this.render(id)
    }

    private render(id: Number): void {
        let self = this
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
                    let ary = json.data as Array<any>
                    let data = ary.map((e) => {
                        let a = JSON.parse(e.payload)
                        let t: string = e.updated_at
                        a.date = t.replace(self._datePtn, "$1 $2")
                        return a
                    })

                    let vals = Object.keys(data[data.length - 1])
                    self.renderGraph(data, vals)
                } else {
                    alert(`データの取得に失敗しました: ${ json.reason }`)
                    window.location.reload(true)
                }
            }
        })
    }

    private renderGraph(jsonAry: Array<any>, values: any): void {
        var chart = c3.generate({
            data: {
                x: 'date',
                xFormat: '%Y-%m-%d %H:%M:%S', // 'xFormat' can be used as custom format of 'x'
                json: jsonAry,
                keys: {
                    x: 'date', // it's possible to specify 'x' when category axis
                    value: values,
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
    }
}

var render = new Render()

var actions = <NodeListOf<HTMLElement>>document.querySelectorAll('.action')
for (let i = 0; i < actions.length; ++i) {
    if (i == 0) {
        let elem = actions[i]
        let name = elem.getAttribute("name")
        if (name == null) {
            console.log("id is null")
        }
        console.log(name)
        let id = Number(name)
        render.GenerateGraph(id)
    }

    actions[i].addEventListener('click', function(e) {
        e.preventDefault()
        let name = this.getAttribute("name")
        if (name == null) {
            console.log("id is null")
            return
        }
        let id = Number(name)
        render.GenerateGraph(id)
    })
}



