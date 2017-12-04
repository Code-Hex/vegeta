import * as request from 'superagent';
import * as c3 from 'c3';
import JSONFormatter from 'json-formatter-js';

enum RenderSpan {
    Week  = "week",
    Month = "month",
    All   = "all"
}

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
        .end(function(err, res) {
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
    
    // page numbers are like these: 0, 1, 2...
    public DataFetch(id: Number, page: Number, span: RenderSpan): Promise<request.Response> {
        return request.post('/mypage/api/data')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({
            tag_id: id,
            page: page,
            span: span,
            limit: 20
        })
    }

    public Graph(prefix: string, rawdata: any[]): void {
        // Get keys to render graph
        let jsonary = rawdata.map((e) => {
            let p = JSON.parse(e.payload)
            let t: string = e.updated_at
            p.date = t.replace(render.datePtn, "$1 $2")
            return p
        })
        let values = Object.keys(jsonary[jsonary.length - 1])

        let chartTag = `#${ prefix }chart`
        let jsonDom = <HTMLInputElement>document.getElementById(`${ prefix }json`)

        this.InitializeDom(prefix)

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

    public InitializeDom(prefix: string): void {
        let chartDom = <HTMLElement>document.getElementById(`${ prefix }chart`)
        while (chartDom.firstChild) {
            chartDom.removeChild(chartDom.firstChild)
        }

        let jsonDom = <HTMLInputElement>document.getElementById(`${ prefix }json`)
        if (jsonDom.childElementCount > 0) {
            jsonDom.removeChild(<Node>jsonDom.firstChild)
        }
    }
}

var render = new Render()
var date = new Date()

var action = <HTMLSelectElement>document.getElementById('action')
var title = <HTMLHtmlElement>document.getElementById('tagname')
var preval = action.value

function GraphWeek(): (response: any) => void {
    return (response: any) => {
        render.InitializeDom('week-')
        let json = response.body
        if (json === undefined) return
        if (!json.is_success) {
            throw new Error(`直近1週間分のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) return
        render.Graph('week-', json.data) // #chart
    }
}

function GraphMonth(): (response: any) => void {
    return (response: any) => {
        render.InitializeDom('month-')
        let json = response.body
        if (json === undefined) return
        if (!json.is_success) {
            throw new Error(`直近1ヶ月分のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) return
        render.Graph('month-', json.data) // #chart
    }
}

function GraphAll(): (response: any) => void {
    return (response: any) => {
        render.InitializeDom('')
        let json = response.body
        if (json === undefined) {
            throw new Error(`全体のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`全体のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            throw new Error(`全体のデータが存在しませんでした`)
        }
        render.Graph('', json.data) // #chart
    }
}

action.addEventListener('change', async (e) => {
    e.preventDefault()
    if (action.value == "") return

    let isCaught = false
    let id = Number(action.value)
    await Promise.all([
        render.DataFetch(id, 0, RenderSpan.Week).then(GraphWeek(), (e) => e),
        render.DataFetch(id, 0, RenderSpan.Month).then(GraphMonth(), (e) => e),
        render.DataFetch(id, 0, RenderSpan.All).then(GraphAll(), (e) => e)
    ]).catch(function(error) {
        alert('タグの切り替え時にエラーが発生しました\n' + error)
        action.value = preval
        isCaught = true
    })

    if (isCaught) {
        let id = Number(preval)
        await Promise.all([
            render.DataFetch(id, 0, RenderSpan.Week).then(GraphWeek(), (e) => e),
            render.DataFetch(id, 0, RenderSpan.Month).then(GraphMonth(), (e) => e),
            render.DataFetch(id, 0, RenderSpan.All).then(GraphAll(), (e) => e)
        ]).catch(function(error) {
            alert('タグの復元中にエラーが発生しました\n' + error)
        })
        return
    }

    // title change
    title.textContent = `タグ${action[action.selectedIndex].text }のグラフ`
    preval = action.value // to restore pull down
})

var addTagElem = <HTMLInputElement>document.getElementById('add-tag')
addTagElem.addEventListener('click', (e) => {
    e.preventDefault()
    render.AddTag()
})

