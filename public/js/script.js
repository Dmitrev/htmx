document.onkeypress = function(e) {
    // Stupid little javascript to fix the number input on Firefox
    // For some reason Firefox allows you to type invalid characters into a number input
    // It will then send an empty string as input to the server
    if (e.target.tagName === 'INPUT' && e.target.type === 'number') {
        if (['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'].includes(e.key) === false) {
            e.preventDefault();
        }
    }
}
