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

    public AddTag(): void {
        let name_input = <HTMLInputElement>document.getElementById('tag_name')
        let name = name_input.value
        request.put('/mypage/api/add_tag')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({ tag_name: name })
        .end(function(err, res){
            if (err || !res.ok) {
                alert('http error: ' + err);
            } else {
                let json = res.body
                if (json.is_success) {
                    alert('タグを追加しました')
                    window.location.reload(true)
                } else {
                    alert(`${ json.reason }`)
                    window.location.reload(true)
                }
            }
        })
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

var action = <HTMLInputElement>document.getElementById('action')
action.addEventListener('change', (e) => {
    e.preventDefault()
    let id = Number(action.value)
    console.log(id)
    render.GenerateGraph(id)
})

var addTagElem = <HTMLInputElement>document.getElementById('add-tag')
console.log(addTagElem)
addTagElem.addEventListener('click', (e) => {
    e.preventDefault()
    render.AddTag()
})

