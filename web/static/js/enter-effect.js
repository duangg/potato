/**
 * Created by Xie on 17/12/15.
 */
function enter(target) {
    if (typeof target != "string") {
        return
    }
    $("#"+target).css("opacity",0);
    $("#"+target).bind("animationend", function () {
        $(target).removeClass("pt-page-fadein");
        $("#"+target).css("opacity",1);
    })
    $("#"+target).addClass("pt-page-fadein");
}
