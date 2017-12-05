import * as request from 'superagent';
import * as c3 from 'c3';
import * as flatpickr from 'flatpickr';
import JSONFormatter from 'json-formatter-js';

enum RenderSpan {
    Week  = "week",
    Month = "month",
    All   = "all"
}

function deepCopy(obj: any): any {
    var copy: any;
    // Handle the 3 simple types, and null or undefined
    if (null == obj || "object" != typeof obj) return obj;
    // Handle Date
    if (obj instanceof Date) {
        copy = new Date();
        copy.setTime(obj.getTime());
        return copy;
    }
    // Handle Array
    if (obj instanceof Array) {
        copy = [];
        for (var i = 0, len = obj.length; i < len; i++) {
            copy[i] = deepCopy(obj[i]);
        }
        return copy;
    }
    // Handle Object
    if (obj instanceof Object) {
        copy = {};
        for (var attr in obj) {
            if (obj.hasOwnProperty(attr)) copy[attr] = deepCopy(obj[attr]);
        }
        return copy;
    }
    throw new Error("Unable to copy obj! Its type isn't supported.");
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
    public DataFetch(id: Number, page: Number, span: RenderSpan, limit: Number): Promise<request.Response> {
        return request.post('/mypage/api/data')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({
            tag_id: id,
            page:   page,
            span:   span,
            limit:  limit
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

var prevWeek: any,
    prevMonth: any,
    prevAll: any

function GraphWeek(): (response: any) => void {
    return (response: any) => {
        render.InitializeDom('week-')
        let json = response.body
        if (json === undefined) return
        if (!json.is_success) {
            throw new Error(`直近1週間分のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) return
        prevWeek = json.data
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
        prevMonth = json.data
        render.Graph('month-', json.data) // #chart
    }
}

function GraphAll(): (response: any) => void {
    return (response: any) => {
        render.InitializeDom('')
        let json = response.body
        if (json === undefined) {
            throw new Error(`全期間のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`全期間のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            throw new Error(`全期間のデータが存在しませんでした`)
        }
        prevAll = json.data
        render.Graph('', json.data) // #chart
    }
}

// Span
var allSpan = <HTMLInputElement>document.querySelector('.calendar')

// Slider
var weekSlider = <HTMLInputElement>document.getElementById('input-week')
var weekValue  = <HTMLElement>document.getElementById('input-week-value')
weekValue.textContent = weekSlider.value
weekSlider.addEventListener('input', (e) => {
    e.preventDefault()
    weekValue.textContent = weekSlider.value
})

var monthSlider = <HTMLInputElement>document.getElementById('input-month')
var monthValue  = <HTMLElement>document.getElementById('input-month-value')
monthValue.textContent = monthSlider.value
monthSlider.addEventListener('input', (e) => {
    e.preventDefault()
    monthValue.textContent = monthSlider.value
})

var allSlider = <HTMLInputElement>document.getElementById('input-all')
var allValue = <HTMLElement>document.getElementById('input-all-value')
allValue.textContent = allSlider.value
allSlider.addEventListener('input', (e) => {
    e.preventDefault()
    allValue.textContent = allSlider.value
})

// Pager
interface PreFetchTrigger {
    button: HTMLButtonElement,
    error: Error,
    data?: any
}

var allReload: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#all-pagination .reload'),
    error: new Error()
}

var allPage: number = 0
var allPrev: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#all-pagination .prev'),
    error: new Error()
}
allPrev.button.disabled = true
var allNext: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#all-pagination .next'),
    error: new Error()
}

allReload.button.addEventListener('mouseover', (e) => {
    e.preventDefault()
    allReload.data = null
    let id        = Number(action.value)
    let alllimit  = Number(allSlider.value)
    render.DataFetch(id, 0, RenderSpan.All, alllimit)
        .then((response: any) => {
            let json = response.body
            if (json === undefined) {
                throw new Error(`更新すべき全期間のデータの取得に失敗しました`)
            }
            if (!json.is_success) {
                throw new Error(`更新すべき全期間のデータの取得に失敗しました: ${ json.reason }`)
            }
            if (json.data.length == 0) {
                throw new Error(`更新すべき全期間のデータが存在しませんでした`)
            }
            allReload.data = json.data
        })
        .catch((e) => {
            allReload.data = null
            allReload.error = e
        })
})

allReload.button.addEventListener('click', (e) => {
    e.preventDefault()

    allPage = 0
    allPrev.button.disabled = true
    allNext.button.disabled = false

    if (allReload.data == null) {
        alert(allReload.error)
        return
    }
    render.Graph('', allReload.data)
    window.scrollTo(0, document.documentElement.scrollTop + 200)
})

allPrev.button.addEventListener('mouseover', (e) => {
    e.preventDefault()
    allPrev.data = null
    let id        = Number(action.value)
    let alllimit  = Number(allSlider.value)
    if (allPage > 0)
        allPrevFetch(id, allPage - 1, RenderSpan.All, alllimit)
})

allPrev.button.addEventListener('click', (e) => {
    e.preventDefault()
    if (allPrev.button.disabled) return

    if (allPage > 0) {
        if (allNext.button.disabled)
            allNext.button.disabled = false
        allPage--
        if (allPage == 0)
            allPrev.button.disabled = true
    }

    if (allPrev.data == null) {
        alert(allPrev.error)
        return
    }

    render.Graph('', allPrev.data)
    window.scrollTo(0, document.documentElement.scrollTop + 200)
    allPrev.data = null

    // Prefetch
    if (allPage > 0) {
        let id        = Number(action.value)
        let alllimit  = Number(allSlider.value)
        allPrevFetch(id, allPage - 1, RenderSpan.All, alllimit)
    }
})

allNext.button.addEventListener('mouseover', (e) => {
    e.preventDefault()
    allNext.data = null
    let id        = Number(action.value)
    let alllimit  = Number(allSlider.value)
    allNextFetch(id, allPage + 1, RenderSpan.All, alllimit)
})

allNext.button.addEventListener('click', (e) => {
    e.preventDefault()
    if (allNext.button.disabled) return
    
    if (allPrev.button.disabled)
        allPrev.button.disabled = false
    allPage++

    if (allNext.data == null) {
        alert(allNext.error)
        return
    }
    
    render.Graph('', allNext.data)
    window.scrollTo(0, document.documentElement.scrollTop + 200)
    allNext.data = null

    // Prefetch
    let id        = Number(action.value)
    let alllimit  = Number(allSlider.value)
    allNextFetch(id, allPage + 1, RenderSpan.All, alllimit)
})

function allPrevFetch(id: Number, page: Number, span: RenderSpan, limit: Number) {
    render.DataFetch(id, page, span, limit)
    .then((response: any) => {
        let json = response.body
        if (json === undefined) {
            throw new Error(`これより以後のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`これより以後のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            allPrev.button.disabled = true
            throw new Error(`これより以後のデータが存在しませんでした`)
        }
        allPrev.data = json.data
        // console.log(allPrev.data)
    })
    .catch((e) => {
        allPrev.data = null
        allPrev.error = e
    })
}

function allNextFetch(id: Number, page: Number, span: RenderSpan, limit: Number) {
    render.DataFetch(id, page, span, limit)
    .then((response: any) => {
        let json = response.body
        if (json === undefined) {
            throw new Error(`これより以前のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`これより以前のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            allNext.button.disabled = true
            throw new Error(`これより以前のデータが存在しませんでした`)
        }
        allNext.data = json.data
        // console.log(allNext.data)
    })
    .catch((e) => {
        allNext.data = null
        allNext.error = e
    })
}

action.addEventListener('change', async (e) => {
    e.preventDefault()
    if (action.value == "") return

    let id         = Number(action.value)
    let weeklimit  = Number(weekSlider.value)
    let monthlimit = Number(monthSlider.value)
    let alllimit   = Number(allSlider.value)

    let isCaught = false

    await Promise.all([
        render.DataFetch(id, 0, RenderSpan.Week, weeklimit).then(GraphWeek(), (e) => e),
        render.DataFetch(id, 0, RenderSpan.Month, monthlimit).then(GraphMonth(), (e) => e),
        render.DataFetch(id, 0, RenderSpan.All, alllimit).then(GraphAll(), (e) => e)
    ]).catch(function(error) {
        alert('タグの切り替え時にエラーが発生しました\n' + error)
        isCaught = true
        action.value = preval
        if (prevWeek != null)  render.Graph('week-', deepCopy(prevWeek))
        if (prevMonth != null) render.Graph('month-', deepCopy(prevMonth))
        if (prevAll != null)   render.Graph('', deepCopy(prevAll))
    })

    if (isCaught) return

    // DatePicker
    if (prevAll.length > 0) {
        flatpickr(allSpan, {
            mode: 'range',
            maxDate: 'today',
            minDate: prevAll[0].updated_at
        })
    }

    // initialize
    allPage = 0
    allNext.button.disabled = false
    allPrev.button.disabled = true

    // title change
    title.textContent = `タグ${action[action.selectedIndex].text }のグラフ`
    preval = action.value // to restore pull down
})

var addTagElem = <HTMLInputElement>document.getElementById('add-tag')
addTagElem.addEventListener('click', (e) => {
    e.preventDefault()
    render.AddTag()
})
