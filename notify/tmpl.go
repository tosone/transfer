package notify

var emailDownloadSuccessTmpl = `<!DOCTYPE html>
<html>

<head>

  <meta charset="utf-8">
  <meta http-equiv="x-ua-compatible" content="ie=edge">
  <title>%s</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style type="text/css">
    @media screen {
      @font-face {
        font-family: 'Merriweather';
        font-style: normal;
        font-weight: 400;
        src: local('Merriweather'), local('Merriweather'), url(http://fonts.gstatic.com/s/merriweather/v8/ZvcMqxEwPfh2qDWBPxn6nmB7wJ9CoPCp9n30ZBThZ1I.woff) format('woff');
      }

      @font-face {
        font-family: 'Merriweather Bold';
        font-style: normal;
        font-weight: 700;
        src: local('Merriweather Bold'), local('Merriweather-Bold'), url(http://fonts.gstatic.com/s/merriweather/v8/ZvcMqxEwPfh2qDWBPxn6nhAPw1J91axKNXP_-QX9CC8.woff) format('woff');
      }
    }

    body,
    table,
    td,
    a {
      -ms-text-size-adjust: 100%%;
      -webkit-text-size-adjust: 100%%;
    }

    img {
      -ms-interpolation-mode: bicubic;
    }

    a[x-apple-data-detectors] {
      font-family: inherit !important;
      font-size: inherit !important;
      font-weight: inherit !important;
      line-height: inherit !important;
      color: inherit !important;
      text-decoration: none !important;
    }

    div[style*="margin: 16px 0;"] {
      margin: 0 !important;
    }

    body {
      width: 100%% !important;
      height: 100%% !important;
      padding: 0 !important;
      margin: 0 !important;
    }

    table {
      border-collapse: collapse !important;
    }

    a {
      color: #CC7953;
    }

    img {
      height: auto;
      line-height: 100%%;
      text-decoration: none;
      border: 0;
      outline: none;
    }
  </style>

</head>

<body style="background-color: #D2C7BA;">
  <div class="preheader"
    style="display: none; max-width: 0; max-height: 0; overflow: hidden; font-size: 1px; line-height: 1px; color: #fff; opacity: 0;">
    A preheader is the short summary text that follows the subject line when an email is viewed in the inbox.
  </div>
  <table border="0" cellpadding="0" cellspacing="0" width="100%%">
    <tr>
      <td align="center" bgcolor="#D2C7BA">
        <table border="0" cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px;">
          <tr>
            <td align="left" bgcolor="#ffffff"
              style="padding: 36px 24px 0; font-family: 'Merriweather Bold', serif; border-top: 5px solid #69BCB1;">
              <h1 style="margin: 0; font-size: 32px; font-weight: 700; letter-spacing: -1px; line-height: 48px;">
                Download File Success</h1>
            </td>
          </tr>
        </table>
      </td>
    </tr>
    <tr>
      <td align="center" bgcolor="#D2C7BA">
        <table border="0" cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px;">
          <tr>
            <td align="left" bgcolor="#ffffff"
              style="padding: 24px; font-family: 'Merriweather', serif; font-size: 16px; line-height: 24px;border-bottom: 5px solid #69BCB1">
              <p style="margin: 0;">File <a href="%s">%s</a> download success.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>

</html>
`
