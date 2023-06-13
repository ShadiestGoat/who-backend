TODO:
- add support for passworded quizes
- add family friendly mode

profile:
- id string primary key
- deadname: []string
- deadlastname: string
- chosenname: []string
- chosenlastname: string
- nickname: string
- order: [3]string (referances to question ids)
- drop_question: string (ref to question id)

- questions strings: {{VARNAME}} <- name
- answers: (case insensitive, max 4)
- name limits: 4

questions:
- id string primary key
- answers []string (max len 4, case insensitive)
- is_multiple_choice bool
- content string

(add password option to not allow folk to share stuff)

Section I: Authentication
1: questions[order[0]] (with replaced "My" or {deadname})
2: questions[order[1]] (with replaced "My" or {deadname})
3: questions[order[3]] (with replaced "My" or {deadname})

Section II: Connecting the dots

order2 := order.filter(not the drop question).sort()

4: questions[order2[0]] (with replaced {nickname})
5: questions[order2[1]] (with replaced {nickname})
6: What is another name for {nickname}?

Section III: Who is she?? Who are they?? Who is he??

7: questions[order2[0]] (with replaced {chosenname})
8: questions[order2[1]] (with replaced {chosenname})
9: What is another name for {chosenname}?

{Character sheet of sorts}

final card: redirect to a pronouns.page
