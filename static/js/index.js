$("#sendbtn").click(function () {
    $.ajax({
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        type: "post",
        url: "/",
        data: JSON.stringify({
            address: $("#address").val(),
            amount: $("#amount").val(),
            token: $("#token").val()
        }),
        success: function (response, status, xhr) {
            if (response["data"]["code"] !== 0) {
                $("#msg").removeClass("text-success").addClass("text-danger").text("Send error: " + response["data"]["message"]);
            } else {
                $("#msg").removeClass("text-danger").addClass("text-success").text("Send successed: " + response["data"]["message"]);
                $(".txlist").before($('<li class="list-group-item"><span class="badge">' + $("#amount").val() + '</span>' + $("#address").val() + '</li>'))
            }
        },
        error: function (response, status, xhr) {
            $("#msg").removeClass("text-success").addClass("text-danger").text("Send error: network disconnected");
        }
    })
});
