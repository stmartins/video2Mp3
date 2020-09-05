function checkField() {
    // alert("entrer une url")
        if (document.getElementById("urlTextField").value == "") {
            console.log("champ url vide")
            return false;
        }
        return true;
}
