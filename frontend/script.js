document.addEventListener("DOMContentLoaded", function() {
    const div = document.querySelector('.room-list');

    const fetchWithTimeout = (url, options, timeout = 7000) => {
        return Promise.race([
            fetch(url, options),
            new Promise((_, reject) =>
                setTimeout(() => reject(new Error('Timeout')), timeout)
            )
        ]);
    };

    // Get the current URL
    const currentURL = window.location.href;

    // Extract the subdomain from the URL
    const subdomain = currentURL.split('.')[0].split('//')[1];

    // Construct the API URL based on the subdomain
    const apiURL = `https://${subdomain}.ch/api/kzu`;

    fetchWithTimeout(apiURL)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(jsonData => {
            console.log('Data:', jsonData);
            
            jsonData.forEach(item => {
                const roomElement = document.createElement('div');
                
                if (item['consecutive'] == 1) {
                    roomElement.className = 'room-container-consecutive';
                } else if (item['canceled'] == 1) {
                    roomElement.className = 'room-container-canceled';
                } else {
                    roomElement.className = 'room-container';
                }
                roomElement.textContent = item['room'];
                
                div.appendChild(roomElement);
            });
        })
        .catch(error => {
            console.error('There was a problem fetching the data:', error);
            
            // Display the error message on the page
            div.textContent = `Error: ${error.message} -- Bitte versuchen Sie es später erneut.`;
            div.className = 'error-message';

            // Send the error details to the server
            fetch(`https://${subdomain}.ch/api/error-logging`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    error: error.message,
                    url: apiURL,
                    timestamp: new Date().toISOString()
                })
            })
            .then(logResponse => {
                if (!logResponse.ok) {
                    console.error('Error reporting failed:', logResponse.statusText);
                }
            })
            .catch(logError => {
                console.error('Failed to send error report:', logError);
            });
        });
});

document.querySelector('.toggle-text').addEventListener('click', function() {
    document.querySelector('.room-list').classList.toggle('show-canceled');
    console.log('toggle canceled');
});

document.querySelector('.slider').addEventListener('click', function() {
    document.querySelector('.room-list').classList.toggle('show-canceled');
    console.log('toggle canceled');
});