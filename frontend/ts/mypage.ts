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

interface FetchParam {
    ID:       Number
    Page:     Number
    Limit:    Number
    Span:     RenderSpan
    StartAt:  string
    EndAt:    string
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
    public DataFetch(param: FetchParam): Promise<request.Response> {
        return request.post('/mypage/api/data')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({
            tag_id:   param.ID,
            page:     param.Page,
            span:     param.Span,
            limit:    param.Limit,
            start_at: param.StartAt,
            end_at:   param.EndAt
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
        let values = Object.keys(jsonary[0])

        let chartTag = `#${ prefix }chart`
        let jsonDom = <HTMLInputElement>document.getElementById(`${ prefix }json`)

        this.InitializeDom(prefix)

        let chart = c3.generate({
            bindto: chartTag,
            data: {
                x: 'date',
                xFormat: '%Y-%m-%d %H:%M:%S', // 'xFormat' can be used as custom format of 'x'
                json: jsonary.reverse(),
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

var prevWeekdata: any,
    prevMonthdata: any,
    prevAlldata: any

function GraphWeek(): (response: any) => void {
    return (response: any) => {
        render.InitializeDom('week-')
        let json = response.body
        if (json === undefined) return
        if (!json.is_success) {
            throw new Error(`直近1週間分のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) return
        prevWeekdata = json.data
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
        prevMonthdata = json.data
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
        prevAlldata = json.data
        render.Graph('', json.data) // #chart
    }
}

// Span
var allSpan = <HTMLInputElement>document.querySelector('.calendar')

function getBetween(): string[] {
    let sp = allSpan.value.split(' to ')
    if (sp.length != 2) return []
    return sp
}

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

//
// WEEK
//

var weekReload: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#week-pagination .reload'),
    error: new Error()
}

var weekPage: number = 0
var weekPrev: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#week-pagination .prev'),
    error: new Error()
}
weekPrev.button.disabled = true
var weekNext: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#week-pagination .next'),
    error: new Error()
}

weekReload.button.addEventListener('mouseover', (e) => {
    e.preventDefault()
    weekReload.data = null
    let id        = Number(action.value)
    let weeklimit  = Number(weekSlider.value)

    render.DataFetch({
        ID:      id,
        Page:    0,
        Span:    RenderSpan.Week,
        Limit:   weeklimit,
        StartAt: '',
        EndAt:   ''
    })
    .then((response: any) => {
        let json = response.body
        if (json === undefined) {
            throw new Error(`更新すべき1週間のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`更新すべき1週間のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            throw new Error(`更新すべき1週間のデータが存在しませんでした`)
        }
        weekReload.data = json.data
    })
    .catch((e) => {
        weekReload.data = null
        weekReload.error = e
    })
})

weekReload.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()

    weekPage = 0
    weekNext.button.disabled = true
    weekPrev.button.disabled = false

    if (weekReload.data == null) {
        alert(weekReload.error)
        return
    }
    render.Graph('week-', weekReload.data)
})

weekNext.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    weekNext.data = null
    prevFetch(weekNext, weekPage - 1, RenderSpan.Week)
})

weekNext.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()
    if (weekNext.button.disabled) return

    if (weekNext.data == null) {
        if (weekNext.error.message != '') {
            alert(weekNext.error)
            return
        }
        // Prefetch
        prevFetch(weekNext, weekPage - 1, RenderSpan.Week)
        return
    }

    if (weekPrev.button.disabled)
        weekPrev.button.disabled = false

    weekPage--

    if (weekPage <= 0) {
        weekPage = 0
        weekNext.button.disabled = true
    }

    render.Graph('week-', weekNext.data)
    weekNext.data = null

    // Prefetch
    prevFetch(weekNext, weekPage - 1, RenderSpan.Week)
})

weekPrev.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    weekPrev.data = null
    nextFetch(weekPrev, weekPage + 1, RenderSpan.Week)
})

weekPrev.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()
    if (weekPrev.button.disabled) return

    if (weekNext.button.disabled)
        weekNext.button.disabled = false

    if (weekPrev.data == null) {
        if (weekPrev.error.message != '') {
            alert(weekPrev.error)
            return
        }
        // Prefetch
        nextFetch(weekPrev, weekPage + 1, RenderSpan.Week)
        return
    }
    weekPage++
    
    render.Graph('week-', weekPrev.data)
    weekPrev.data = null

    // Prefetch
    nextFetch(weekPrev, weekPage + 1, RenderSpan.Week)
})

//
// MONTH
//

var monthReload: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#month-pagination .reload'),
    error: new Error()
}

var monthPage: number = 0
var monthPrev: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#month-pagination .prev'),
    error: new Error()
}
monthPrev.button.disabled = true
var monthNext: PreFetchTrigger = {
    button: <HTMLButtonElement>document.querySelector('#month-pagination .next'),
    error: new Error()
}

monthReload.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    monthReload.data = null
    let id        = Number(action.value)
    let monthlimit  = Number(monthSlider.value)

    render.DataFetch({
        ID:      id,
        Page:    0,
        Span:    RenderSpan.Month,
        Limit:   monthlimit,
        StartAt: '',
        EndAt:   ''
    })
    .then((response: any) => {
        let json = response.body
        if (json === undefined) {
            throw new Error(`更新すべき1ヶ月のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`更新すべき1ヶ月のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            throw new Error(`更新すべき1ヶ月のデータが存在しませんでした`)
        }
        monthReload.data = json.data
    })
    .catch((e) => {
        monthReload.data = null
        monthReload.error = e
    })
})

monthReload.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()

    monthPage = 0
    monthNext.button.disabled = true
    monthPrev.button.disabled = false

    if (monthReload.data == null) {
        alert(monthReload.error)
        return
    }
    render.Graph('month-', monthReload.data)
})

monthNext.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    monthNext.data = null
    prevFetch(monthNext, monthPage - 1, RenderSpan.Month)
})

monthNext.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()
    if (monthNext.button.disabled) return

    if (monthNext.data == null) {
        if (monthNext.error.message != '') {
            alert(monthNext.error)
            return
        }
        // Prefetch
        prevFetch(monthNext, monthPage - 1, RenderSpan.Month)
        return
    }

    if (monthPrev.button.disabled)
        monthPrev.button.disabled = false

    monthPage--

    if (monthPage <= 0) {
        monthPage = 0
        monthNext.button.disabled = true
    }

    render.Graph('month-', monthNext.data)
    monthNext.data = null

    // Prefetch
    prevFetch(monthNext, monthPage - 1, RenderSpan.Month)
})

monthPrev.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    monthPrev.data = null
    nextFetch(monthPrev, monthPage + 1, RenderSpan.Month)
})

monthPrev.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()
    if (monthPrev.button.disabled) return

    if (monthNext.button.disabled)
        monthNext.button.disabled = false

    if (monthPrev.data == null) {
        if (monthPrev.error.message != '') {
            alert(monthPrev.error)
            return
        }
        // Prefetch
        nextFetch(monthPrev, monthPage + 1, RenderSpan.Month)
        return
    }
    monthPage++
    
    render.Graph('month-', monthPrev.data)
    monthPrev.data = null

    // Prefetch
    nextFetch(monthPrev, monthPage + 1, RenderSpan.Month)
})

//
// ALL
//

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
    e.stopImmediatePropagation()
    allReload.data = null
    let id        = Number(action.value)
    let alllimit  = Number(allSlider.value)

    let calendar  = allSpan.value.split(' to ')
    let start_at = calendar[0]
    let end_at   = calendar[1] || ""
    render.DataFetch({
        ID:      id,
        Page:    0,
        Span:    RenderSpan.All,
        Limit:   alllimit,
        StartAt: start_at,
        EndAt:   end_at
    })
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
    e.stopImmediatePropagation()

    allPage = 0
    allNext.button.disabled = true
    allPrev.button.disabled = false

    if (allReload.data == null) {
        alert(allReload.error)
        return
    }
    render.Graph('', allReload.data)
})

allNext.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    allNext.data = null
    prevFetch(allNext, allPage - 1, RenderSpan.All)
})

allNext.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()
    if (allNext.button.disabled) return

    if (allNext.data == null) {
        if (allNext.error.message != '') {
            alert(allNext.error)
            return
        }
        // Prefetch
        prevFetch(allNext, allPage - 1, RenderSpan.All)
        return
    }

    if (allPrev.button.disabled)
        allPrev.button.disabled = false

    allPage--

    if (allPage <= 0) {
        allPage = 0
        allNext.button.disabled = true
    }

    render.Graph('', allNext.data)
    allNext.data = null

    // Prefetch
    prevFetch(allNext, allPage - 1, RenderSpan.All)
})

allPrev.button.addEventListener('mouseover', (e) => {
    e.stopImmediatePropagation()
    allPrev.data = null
    nextFetch(allPrev, allPage + 1, RenderSpan.All)
})

allPrev.button.addEventListener('click', (e) => {
    e.stopImmediatePropagation()
    if (allPrev.button.disabled) return

    if (allNext.button.disabled)
        allNext.button.disabled = false

    if (allPrev.data == null) {
        if (allPrev.error.message != '') {
            alert(allPrev.error)
            return
        }
        // Prefetch
        nextFetch(allPrev, allPage + 1, RenderSpan.All)
        return
    }
    allPage++
    
    render.Graph('', allPrev.data)
    allPrev.data = null

    // Prefetch
    nextFetch(allPrev, allPage + 1, RenderSpan.All)
})

function prevFetch(prev: PreFetchTrigger, page: Number, span: RenderSpan) {
    let between: string[] = ['', '']
    if (span == RenderSpan.All) {
        between = getBetween()
    }
    if (page < 0) page = 0

    let id    = Number(action.value)
    let limit = Number(allSlider.value)

    render.DataFetch({
        ID:      id,
        Page:    page,
        Span:    span,
        Limit:   limit,
        StartAt: between[0],
        EndAt:   between[1]
    })
    .then((response: any) => {
        let json = response.body
        if (json === undefined) {
            throw new Error(`これより以後のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`これより以後のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            prev.button.disabled = true
            throw new Error(`これより以後のデータが存在しませんでした`)
        }
        prev.data = json.data
    })
    .catch((e) => {
        prev.data = null
        prev.error = e
    })
}

function nextFetch(next: PreFetchTrigger, page: Number, span: RenderSpan) {
    let between: string[] = ['', '']
    if (span == RenderSpan.All) {
        between = getBetween()
    }

    let id    = Number(action.value)
    let limit = Number(allSlider.value)

    render.DataFetch({
        ID:      id,
        Page:    page,
        Span:    span,
        Limit:   limit,
        StartAt: between[0],
        EndAt:   between[1],
    })
    .then((response: any) => {
        let json = response.body
        if (json === undefined) {
            throw new Error(`これより以前のデータの取得に失敗しました`)
        }
        if (!json.is_success) {
            throw new Error(`これより以前のデータの取得に失敗しました: ${ json.reason }`)
        }
        if (json.data.length == 0) {
            next.button.disabled = true
            return
        }
        next.data = json.data
    })
    .catch((e) => {
        next.data = null
        next.error = e
    })
}

action.addEventListener('change', async (e) => {
    e.preventDefault()
    if (action.value == "") return

    let id         = Number(action.value)
    let weeklimit  = Number(weekSlider.value)
    let monthlimit = Number(monthSlider.value)
    let alllimit   = Number(allSlider.value)

    let calendar  = allSpan.value.split(' to ')
    let start_at: string = calendar[0]
    let end_at: string   = calendar[1] || ''

    let isCaught = false

    await Promise.all([
        render.DataFetch({
            ID:      id,
            Page:    0,
            Span:    RenderSpan.Week,
            Limit:   weeklimit,
            StartAt: '',
            EndAt:   ''
        }).then(GraphWeek(), (e) => e),
        render.DataFetch({
            ID:      id,
            Page:    0,
            Span:    RenderSpan.Month,
            Limit:   monthlimit,
            StartAt: '',
            EndAt:   ''
        }).then(GraphMonth(), (e) => e),
        render.DataFetch({
            ID:      id,
            Page:    0,
            Span:    RenderSpan.All,
            Limit:   alllimit,
            StartAt: start_at,
            EndAt:   end_at,
        }).then(GraphAll(), (e) => e)
    ]).catch(function(error) {
        alert('タグの切り替え時にエラーが発生しました\n' + error)
        isCaught = true
        action.value = preval
        if (prevWeekdata != null)  render.Graph('week-', deepCopy(prevWeekdata))
        if (prevMonthdata != null) render.Graph('month-', deepCopy(prevMonthdata))
        if (prevAlldata != null)   render.Graph('', deepCopy(prevAlldata))
    })

    if (isCaught) return

    // DatePicker
    if (prevAlldata != null && prevAlldata.length > 0) {
        flatpickr(allSpan, {
            mode: 'range',
            maxDate: prevAlldata[0].updated_at
        })
    }

    // initialize
    allPage = 0
    weekPrev.button.disabled = false
    weekNext.button.disabled = true
    monthPrev.button.disabled = false
    monthNext.button.disabled = true
    allPrev.button.disabled = false
    allNext.button.disabled = true

    // title change
    title.textContent = `タグ${action[action.selectedIndex].text }のグラフ`
    preval = action.value // to restore pull down
})

var addTagElem = <HTMLInputElement>document.getElementById('add-tag')
addTagElem.addEventListener('click', (e) => {
    e.preventDefault()
    render.AddTag()
})
