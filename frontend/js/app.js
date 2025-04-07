$(document).ready(function() {
    // Setup CORS for cross-domain requests
    $.ajaxSetup({
        crossDomain: true,
        xhrFields: {
            withCredentials: false
        }
    });
    
    // Check authentication status when the page loads
    checkAuthStatus();
});