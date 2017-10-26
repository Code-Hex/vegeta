import * as request from 'superagent';
import * as c3 from 'c3';
import JSONFormatter from 'json-formatter-js';

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
    public get datePtn(): RegExp {
        return this._datePtn
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

    public DataFetch(id: Number): Promise<request.Response> {
        let self = this
        return request.post('/mypage/api/data')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({ tag_id: id })
    }

    public Graph(prefix: string, jsonary: any[], values: any, rawdata: any[]): void {
        let chartTag = `#${ prefix }chart`
        let jsonTag = `${ prefix }json`
        // iniialize
        let jsonDom = <HTMLInputElement>document.getElementById(jsonTag)
        if (jsonDom.childElementCount > 0) {
            jsonDom.removeChild(<Node>jsonDom.firstChild)
        }

        let chart = c3.generate({
            bindto: chartTag,
            data: {
                x: 'date',
                xFormat: '%Y-%m-%d %H:%M:%S', // 'xFormat' can be used as custom format of 'x'
                json: jsonary,
                keys: {
                    x: 'date', // it's possible to specify 'x' when category axis
                    value: values,
                },
                onmouseover: function (d) {
                    let raw = rawdata[d.index]
                    raw.payload = jsonary[d.index]
                    delete raw.updated_at
                    if (jsonDom.childElementCount > 0) {
                        jsonDom.removeChild(<Node>jsonDom.firstChild)
                    }
                    let formatter = new JSONFormatter(raw, Infinity)
                    jsonDom.appendChild(formatter.render())
                },
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
var date = new Date()

var action = <HTMLSelectElement>document.getElementById('action')
var title = <HTMLHtmlElement>document.getElementById('tagname')
var preval = action.value

action.addEventListener('change', (e) => {
    e.preventDefault()
    if (action.value == "") {
        return
    }
    let id = Number(action.value)
    render.DataFetch(id)
    .then(function(response) {
        let json = response.body
        if (json === undefined) {
            alert(`データの取得に失敗しました`)
            action.value = preval
            return
        }
        if (!json.is_success) {
            alert(`データの取得に失敗しました: ${ json.reason }`)
            action.value = preval
            return
        }
        console.log(json)
        if (json.data.length == 0) {
            alert(`データが存在しませんでした`)
            action.value = preval
            return
        }

        let ary: any[] = json.data
        let data = ary.map((e) => {
            let p = JSON.parse(e.payload)
            let t: string = e.updated_at
            p.date = t.replace(render.datePtn, "$1 $2")
            return p
        })

        // get data in a month
        let inMonth = new Date()
        inMonth.setMonth(date.getMonth() - 1)
        console.log(inMonth)
        let inMonthAry = ary.filter((v) => new Date(v.updated_at).getTime() >= inMonth.getTime() )
        let inMonthData = data.filter((v) => new Date(v.date).getTime() >= inMonth.getTime() )

        // get data in a week
        let inWeek = new Date()
        inWeek.setDate(date.getDate() - 7)
        console.log(inWeek)
        let inWeekAry = inMonthAry.filter((v) => new Date(v.updated_at).getTime() >= inWeek.getTime() )
        let inWeekData = inMonthData.filter((v) => new Date(v.date).getTime() >= inWeek.getTime() )

        // Get keys to render graph
        let vals = Object.keys(data[data.length - 1])

        // render
        render.Graph('week-', inWeekData, vals, inWeekAry)    // #week-chart
        render.Graph('month-', inMonthData, vals, inMonthAry) // #month-chart
        render.Graph('', data, vals, ary)                     // #chart

        // title change
        title.textContent = `タグ${action[action.selectedIndex].text }のグラフ`

        // to restore pull down
        preval = action.value
    }, function(error) {
        alert('http error: ' + error)
        action.value = preval
    })
})

var addTagElem = <HTMLInputElement>document.getElementById('add-tag')
addTagElem.addEventListener('click', (e) => {
    e.preventDefault()
    render.AddTag()
})

