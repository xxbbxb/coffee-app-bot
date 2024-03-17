Command/TextPrefix - относятся к Mux, не могут быть вложены в Route
Route - инъекция роутера/поддерева хэндлеров
CallbackData - прямой вызов хэндлера func (Update) {} при совпадении pattern-а
Default - хэндлер по умолчанию когда ничего не отработало

// не роутер, не часть либы можно цеплять в Mux.default
Expect - Mux default handler обрабатывается в самом конце если роуты или cb не сматичились (тогда предыдущие меню и команды не будут заблокированы). Если Mux уже ожидает что-то то повторный вызов Expect приведет к сбросу ожидания


r.Use(Auth(aclFile))
r.TextPrefix("/start", func(u *Update) {
	x := tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("about", "/about"),
        tgbotapi.NewInlineKeyboardButtonData("skills", "/skills"),
		tgbotapi.NewInlineKeyboardButtonData("contacts", "/contacts"),
		tgbotapi.NewInlineKeyboardButtonData("search", "/search"),
    )
	u.Bot.Send(x)
})
r.Route("/about", func (r Routable) {
	r.CallbackDataHandler("/", func (u *Update) {

	})
	r.CallbackDataHandler("/edit", func (u *Update) {

	})
	r.CallbackDataHandler("/save", func (u *Update) {

	})
})
r.CallbackDataHandler("/skills", func(u *Update) {
 
})
r.CallbackDataHandler("/contacts", func(u *Update) {

})
r.CallbackDataHandler("/search", func(u *Update) {
	r.Expect("/search/input", func (u *Update) {

	}, cancelFunc ())
})


// TODO: Add empty subRouter check or root default