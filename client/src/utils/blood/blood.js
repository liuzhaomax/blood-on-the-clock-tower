export const blood = () => {
    var h = 200, w = 600
    var canvas = document.createElement("canvas")
    var ctx = canvas.getContext("2d")
    document.body.appendChild(canvas)
    canvas.height = h
    canvas.width = w
    canvas.style.position = "absolute"
    canvas.style.top = "0"

    var Blood = function(x, y, r, s) {
        this.x = x
        this.y = y
        this.r = r
        this.cy = y
        this.speed = s
    }

    Blood.prototype.draw = function() {
        ctx.beginPath()
        ctx.arc(this.x + 1, this.cy, this.r, 0, Math.PI * 2)
        ctx.closePath()
        ctx.fillStyle = "red"
        ctx.fill()
    }

    var bloodDrops = [
        new Blood(10, 70, 1, 2),
        new Blood(35, 32, 1, 1.5),
        new Blood(53, 72, 1.5, 1),
        new Blood(80, 74, 2, 1.7),
        new Blood(100, 68, 1, 1.3),
        new Blood(154, 71, 1, 1.9),
        new Blood(174, 71, 1, 1.4),
        new Blood(222, 76, 1.3, 1.1),
        new Blood(263, 72, 1.8, 0.7),
        new Blood(280, 75, 1.4, 1.3),
        new Blood(325, 72, 1.8, 1.2),
        new Blood(380, 75, 1.4, 0.9),
        new Blood(395, 88, 0.8, 1.3),
        new Blood(418, 70, 1.2, 1.6),
        new Blood(466, 67, 0.8, 1.3),
        new Blood(487, 71, 1.2, 1.2),
        new Blood(512, 74, 1.7, 1.5),
        new Blood(542, 74, 1.7, 0.8)
    ]

    var loop = function() {
        ctx.fillStyle = "rgba(0,0,0,0.005)"
        ctx.fillRect(0, 0, w, h)
        for (var i = 0; i < bloodDrops.length; ++i) {
            if (Math.random() > 0.5) bloodDrops[i].cy += bloodDrops[i].speed
            else bloodDrops[i].cy += bloodDrops[i].speed / 3
            if (bloodDrops[i].cy > h)
                bloodDrops[i].cy = bloodDrops[i].y
            bloodDrops[i].draw()
        }
        requestAnimationFrame(loop)
    }

    loop()
}