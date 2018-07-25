$("#sendbtn").click(function () {
    $.ajax({
        type: "post",
        url: "/",
        data: {
            "address": $("#address").val(),
            "amount": $("#amount").val()
        },
        success: function (response, status, xhr) {
            if (response["Code"] != 0) {
                $("#msg").removeClass("text-success").addClass("text-danger").text("Send error: " + response["Message"]);
            } else {
                $("#msg").removeClass("text-danger").addClass("text-success").text("Send successed: " + response["Message"]);
                $(".txlist").before($('<li class="list-group-item"><span class="badge">'+ $("#amount").val() + '</span>' + $("#address").val() + '</li>'))
            }
        },
        error: function (response, status, xhr) {
            $("#msg").removeClass("text-success").addClass("text-danger").text("Send error: network disconnected");
            $(".txlist").before($('<li class="list-group-item"><span class="badge">'+ $("#amount").val() + '</span>' + $("#address").val() + '</li>'))
        }
    })
});
