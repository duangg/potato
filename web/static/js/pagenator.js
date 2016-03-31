/**
 * Created by Xie on 17/12/15.
 */
function pagenator(target, callback) {
    this.currActivePage = 0;
    this.pn = 0;
    this.handler = target;
    this.cb = callback;
}
pagenator.prototype._pageSwitch = function (pn) {
    if (pn == this.currActivePage)
        return;
    $("#" + this.handler).children("[pageid='" + this.currActivePage + "']").removeClass("active");
    $("#" + this.handler).children("[pageid='" + pn + "']").addClass("active");
    this.currActivePage = pn;
};
pagenator.prototype.Init = function (pagenu) {
    if (arguments.length == 0)
        pagenu = 1;
    console.log($("#" + this.handler).children());
    $("#" + this.handler).children().remove();
    for (var i = 0; i < pagenu; i++) {
        console.log("add");
        this.addPage();
    }
    $("#" + this.handler).children("[pageid='0']").addClass("active");
    $("#" + this.handler).prepend("<li pageid='page-prev'><a href='#'>&laquo;</a></li>");
    $("#" + this.handler).append("<li pageid='page-next'><a href='#'>&raquo;</a></li>");
    $("#" + this.handler).children("[pageid='page-prev']").click(function () {
        if (this.currActivePage <= 0)
            return;
        this._pageSwitch(this.currActivePage - 1);
        this.cb(this.currActivePage);
    }.bind(this));
    $("#" + this.handler).children("[pageid='page-next']").click(function () {
        if (this.currActivePage >= this.pn - 1)
            return;
        this._pageSwitch(this.currActivePage + 1);
        this.cb(this.currActivePage);
    }.bind(this));

};
pagenator.prototype.addPage = function () {
    var txt = "<li pageid=" + this.pn + "><a href='#'>" + (this.pn + 1) + "</a></li>";
    if (this.pn == 0) {
        $("#" + this.handler).append(txt);
    } else {
        $("#" + this.handler).children("[pageid='" + (this.pn - 1) + "']").after(txt);
    }
    $("#" + this.handler).children("[pageid='" + this.pn + "']").click(function (data) {
        this._pageSwitch(this.getPage(data.currentTarget));
        this.cb(this.currActivePage);
    }.bind(this));
    this.pn++;
};
pagenator.prototype.removePage = function () {
    if (this.pn <= 0)
        return;
    $("#" + this.handler).children("[pageid='" + (this.pn - 1) + "']").remove();
    this.pn--;
};
pagenator.prototype.getPage = function (target) {
    return parseInt(target.innerText) - 1;
};