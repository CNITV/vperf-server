# Performante Vianiste Server

## API

Dacă nu este menționat altfel, răspunsul este gol. In caz de eroare,
răspunsul conține eroarea și codul HTTP corespunzător.

Toate căile care încep cu `/admin` necesită autentificare.

### POST `/admin/start`

Pornește concursul.

### POST `/admin/pause?reason=string`

Suspendă concursul. Parametrul `reason` conține motivul suspendării și este
opțional.

### POST `/admin/stop`

Termină concursul.

### PUT `/admin/team/{id}/special`

In body primește un singur întreg, numărul problemei care va fi setată ca
specială.

Setează pentru echipa `{id}` problema data ca fiind specială.

### POST `/admin/team/{id}/submit/{problem_no}`

In body primește un singur întreg, răspunsul la problemă.

Trimite răspunsul la problema `{problem_no}` din partea echipei `{id}`.

### POST `/admin/team/{id}/fine`

In body primește un singur întreg, numărul de puncte pentru sancțiune.

Sancționează echipa `{id}` cu un număr de puncte.

### DELETE `/admin/team/{id}?forget=bool`

Descalifică ireversibil echipa din concurs. Daca parametrul `forget` este setat
cu valoarea `true`, atunci se consideră ca echipa nu a participat deloc.

### GET `/admin/log`

Returnează jurnalul evenimentelor.

### DELETE `/admin/log/{id}`

Șterge intrarea `{id}` din jurnal și recalculează scorurile.


### GET `/status`

Returnează starea concursului.

