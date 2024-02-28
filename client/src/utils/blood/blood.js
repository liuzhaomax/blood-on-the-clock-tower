import {sleep} from "../time"

export const blood = async () => {
    await sleep(3000)
    var h = 180, w = 345
    var canvas = document.createElement("canvas")
    var ctx = canvas.getContext("2d")
    let container = document.getElementById("Title-wrap")
    container.appendChild(canvas)
    canvas.height = h
    canvas.width = w

    var Blood = function(x, y, r, s) {
        this.x = x
        this.y = y
        this.r = r
        this.cy = y
        this.speed = s
        this.opacity = 1 // 初始透明度为1，完全不透明
    }

    Blood.prototype.draw = function() {
        ctx.globalAlpha = this.opacity // 设置圆的透明度
        ctx.beginPath()
        ctx.arc(this.x + 1, this.cy, this.r, 0, Math.PI * 2)
        ctx.closePath()
        ctx.fillStyle = "rgba(160, 42, 42, 0.5)" // 半透明的红色
        // ctx.fillStyle = "brown"
        ctx.fill()
    }

    var bloodDrops = [
        // 血
        new Blood(13, 76, 0.7, 0.9),
        new Blood(31, 72, 1, 1.2),
        new Blood(38, 30, 1.3, 1.5),
        new Blood(70, 38, 1.2, 1),
        new Blood(80, 74, 1.6, 1.7),
        // 染
        new Blood(100, 68, 1, 1.3),
        new Blood(115, 40, 0.8, 0.8),
        new Blood(135, 75, 1, 1.3),
        new Blood(160, 62, 1.5, 1.9),
        // 钟
        new Blood(182, 50, 0.8, 0.9),
        new Blood(203, 72, 1, 1.4),
        new Blood(220, 55, 0.7, 0.8),
        new Blood(230, 82, 1, 1.1),
        new Blood(250, 40, 1.3, 0.9),
        // 楼
        new Blood(268, 72, 1.3, 0.7),
        new Blood(280, 82, 1.2, 1.3),
        new Blood(300, 80, 0.8, 1),
        new Blood(334, 62, 1.8, 1.2),
    ]

    let stop = false

    var loop = async () => {
        // ctx.fillStyle = "rgba(0,0,0,0.005)"
        ctx.fillStyle = "rgba(0,0,0,0)"
        ctx.fillRect(0, 0, w, h)

        // for (var i = 0; i < bloodDrops.length; ++i) {
        //     if (Math.random() > 0.5) bloodDrops[i].cy += bloodDrops[i].speed
        //     else bloodDrops[i].cy += bloodDrops[i].speed / 3
        //     if (bloodDrops[i].cy > h)
        //         bloodDrops[i].cy = bloodDrops[i].y
        //     bloodDrops[i].draw()
        // }

        for (var i = 0; i < bloodDrops.length; ++i) {
            bloodDrops[i].opacity -= 0.006 // 透明度降低速度
            bloodDrops[i].r -= 0.005 // 半径减小
            if (bloodDrops[i].r <= 0) {
                stop = true
                break
            }
            if (Math.random() > 0.5) bloodDrops[i].cy += bloodDrops[i].speed
            else bloodDrops[i].cy += bloodDrops[i].speed / 3
            if (bloodDrops[i].cy > h) bloodDrops[i].cy = bloodDrops[i].y
            // if (bloodDrops[i].opacity <= 0) {
            //     bloodDrops[i].opacity = 1 // 重置血滴完全不透明
            // }
            bloodDrops[i].cy += bloodDrops[i].speed
            bloodDrops[i].draw()

        }
        if (!stop) {
            requestAnimationFrame(loop)
        }
        // if (!canvas.classList.contains("bottom-line")) {
        //     await sleep(2000)
        //     canvas.classList.add("bottom-line")
        // }
    }

    loop()
}