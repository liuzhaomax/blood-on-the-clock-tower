// 要操作的元素
// let box=document.querySelector(".box"),
//     titbox=document.querySelector(".tit-box"),
//     fls=document.querySelectorAll(".fl"),
//     radios=document.querySelectorAll(".radio")

// 要点击的词语
let arr=["一帆风顺","二龙腾飞","三羊开泰","四季平安","五福临门"]
// 混淆的字
let dic="一乙二十丁厂七卜人入八九几儿了力乃刀又三于干亏士工土才寸下大丈与万上小口巾山千乞川亿个勺久凡及夕丸么广亡门义之尸弓己已子卫也女飞刃习叉马乡丰王井开夫天无元专云扎艺木五支厅不太犬区历尤友匹车巨牙屯比互切瓦止少日中冈贝内水见午牛手毛气升长仁什片仆化仇币仍仅斤爪反介父从今凶分乏公仓月氏勿欠风丹匀乌凤勾文六方火为斗忆订计户认心尺引丑巴孔队办以允予劝双书幻玉刊示末未击打巧正扑扒功扔去甘世古节本术可丙左厉右石布龙平灭轧东卡北占业旧帅归且旦目叶甲申叮电号田由史只央兄叼叫另叨叹四生失禾丘付仗代仙们仪白仔他斥瓜乎丛令用甩印乐句匆册犯外处冬鸟务包饥主市立闪兰半汁汇头汉宁穴它讨写让礼训必议讯记永司尼民出辽奶奴加召皮边发孕圣对台矛纠母幼丝式刑动扛寺吉扣考托老执巩圾扩扫地扬场耳共芒亚芝朽朴机权过臣再协西压厌在有百存而页匠夸夺灰达列死成夹轨邪划迈毕至此贞师尘尖劣光当早吐吓虫曲团同吊吃因吸吗屿帆岁回岂刚则肉网年朱先丢舌竹迁乔伟传乒乓休伍伏优伐延件任伤价份华仰限妹姑姐姓始驾参艰线练组细驶织终驻驼绍经贯奏春帮珍玻毒型挂封帘实试郎诗肩房诚衬衫视话诞询该详建肃录隶居届刷屈弦承孟孤陕降河沾泪油泊沿泡注泻泳泥沸波泼泽治怖性怕怜怪学宝宗定宜审宙官空京享店夜庙府底剂郊废净盲放刻育闸闹郑券卷单炒炊炕炎炉沫浅法泄斧爸采受乳贪念贫肤肺肢肿胀朋股肥服胁周昏鱼兔狐忽狗备饰饱饲变"

// 清空重置元素
export const clearCaptcha = (fls, radios) => {
    fls.forEach((fl)=>{
        fl.style.display="none"
    })
    radios.forEach((radio)=>{
        radio.style.opacity=0
    })
}

export const startCaptcha = (box, titbox, fls, radios, container, onSuccess) => {
    // 初始化box
    box.innerHTML="<div class='bg-blur'></div>"
    // 初始化titbox
    titbox.innerHTML=""
    titbox.style.color="#333"

    // 获取要点击的词语
    let math=arr[Math.floor(Math.random() * arr.length)]
    titbox.innerHTML=`请依次点击: <span>${math}</span>`

    let timer
    let place=[
        {left:"120px",top:"0px"},
        {left:"160px",top:"20px"},
        {left:"80px",top:"200px"},
        {left:"240px",top:"260px"},
        {left:"40px",top:"260px"},
        {left:"240px",top:"50px"},
        {left:"40px",top:"150px"},
        {left:"40px",top:"110px"},
        {left:"160px",top:"120px"},
    ]
    place.sort(()=>{
        return Math.random() - 0.5
    })

    let verify=[]
    let extra_str=""
    for(let i=0;i<5;i++){
        extra_str+=dic[Math.floor(Math.random() * dic.length)]
    }
    extra_str=math+extra_str
    // 分别设置每个字的位置,属性
    for(let i in extra_str){
        verify.push(i)
        let left=Math.floor(Math.random() * (i*10))
        let top=Math.floor(Math.random() * (i*10))

        let divs=document.createElement("div")
        divs.classList.add("fl")
        divs.style.left=place[i].left
        divs.style.top=place[i].top

        let span=document.createElement("span")
        span.innerText=extra_str[i]
        span.style.left=left+"px"
        span.style.top=top+"px"

        span.setAttribute("data-index",i)
        span.setAttribute("data-judge",true)
        divs.appendChild(span)
        box.appendChild(divs)
    }

    let cspan=[]
    let fspans=box.querySelectorAll(".fl span")
    fspans.forEach((fspan)=>{
        // 为每个字绑定点击事件
        fspan.addEventListener("click",function(){
            if(fspan.getAttribute("data-judge")){
                cspan.push(fspan.getAttribute("data-index"))
                fspan.setAttribute("data-judge",false)
            }else{
                console.log("重复点击")
            }
        })
    })

    let a=0
    // 为box绑定点击事件
    box.addEventListener("click",function(event){
        // 小绿点补偿计算
        let leftDistance = container.getBoundingClientRect().left
        let topDistance = container.getBoundingClientRect().top
        // 小绿点(点击显示序号)
        a++
        let rad=document.createElement("div")
        rad.classList.add("radio")
        rad.innerText=a
        rad.style.left=event.pageX - box.offsetLeft -leftDistance + "px"
        rad.style.top=event.pageY - box.offsetTop+20 -topDistance + "px"
        box.appendChild(rad)
        // 小绿点 结束
        // 共四个字
        if(a===4){
            if(cspan.join()===verify.slice(0,4).join()){
                titbox.innerHTML="验证成功"
                titbox.style.color="green"
                clearCaptcha(fls, radios)
                onSuccess(true)
            }else{
                titbox.innerHTML="验证失败"
                titbox.style.color="red"
                clearCaptcha(fls, radios)
                onSuccess(false)
            }
            // 清除定时器
            clearTimeout(timer)
            // 1秒后执行start
            timer=setTimeout(() => {
                startCaptcha(box, titbox, fls, radios, container)
            }, 1000)
            // 重置a
            a=0
        }
    })
}
