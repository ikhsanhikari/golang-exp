package controller

// func (h *handler) handleGetRelatedArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var (
// 		project, _ = authpassport.GetProject(r)
// 		pid        = project.ID
// 		_id        = ps.ByName("id")
// 		id, err    = strconv.ParseInt(_id, 10, 64)
// 		resVid     responseVideo
// 		params     getRelated
// 		limit      = 4 // default limit related
// 	)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	err = form.BindFlag(form.Bvalidate, &params, r)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	if params.Limit != 0 {
// 		limit = params.Limit
// 	}

// 	article, err := h.articles.Get(id, pid)
// 	if err == sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Article not found", http.StatusNotFound)
// 		return
// 	}
// 	if err != nil && err != sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get article", http.StatusInternalServerError)
// 		return
// 	}

// 	als, err := h.articleLists.SelectByArticle(article.ID, pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 		return
// 	}

// 	listID := als[0].ListID
// 	als, err = h.articleLists.SelectByList(listID, pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get articles", http.StatusInternalServerError)
// 		return
// 	}

// 	relatedArticles := make([]view.DataResponse, 0, len(als))
// 	if len(als) > 0 {
// 		ids := make([]int64, 0, len(als))
// 		for _, al := range als {
// 			if al.ArticleID != id {
// 				ids = append(ids, al.ArticleID)
// 			}
// 		}

// 		articles, err := h.articles.SelectByIDs(ids, pid, limit)
// 		if err != nil {
// 			log.Println(err)
// 			view.RenderJSONError(w, "Failed get articles", http.StatusInternalServerError)
// 			return
// 		}

// 		for i := range articles {
// 			als, err := h.articleLists.SelectByArticle(articles[i].ID, pid)
// 			if err != nil {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 				return
// 			}

// 			ids := make([]int64, 0, len(als))
// 			for _, al := range als {
// 				ids = append(ids, al.ListID)
// 			}

// 			lists, err := h.lists.SelectByIDs(ids, pid)
// 			if err != nil {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 				return
// 			}

// 			names := make([]string, 0, len(lists))
// 			for _, list := range lists {
// 				names = append(names, list.Name)
// 			}

// 			tag := make([]string, 0, 10)
// 			if articles[i].Tags.Valid && articles[i].Tags.String != "" {
// 				tag = strings.Split(articles[i].Tags.String, ",")
// 			}

// 			metaKeyword := make([]string, 0, 10)
// 			if articles[i].MetaKeywords.Valid && articles[i].MetaKeywords.String != "" {
// 				metaKeyword = strings.Split(articles[i].MetaKeywords.String, ",")
// 			}

// 			var video *videoData
// 			if articles[i].VideoID.Valid && articles[i].VideoID.String != "" {
// 				URL := "https://supersoccer.tv/api/v2/videos/" + articles[i].VideoID.String
// 				request, _ := http.NewRequest("GET", URL, nil)
// 				response, err := h.client.Do(request)
// 				if err == nil && response.StatusCode == 200 {
// 					body, _ := ioutil.ReadAll(response.Body)
// 					_ = json.Unmarshal(body, &resVid)
// 					if len(resVid.Data) > 0 {
// 						v := resVid.Data[0]
// 						video = &v
// 					}
// 				}
// 			}

// 			relatedArticles = append(relatedArticles, view.DataResponse{
// 				Type: "articles",
// 				ID:   articles[i].ID,
// 				Attributes: view.ArticleAttributes{
// 					Title:           articles[i].Title,
// 					Author:          articles[i].Author.Ptr(),
// 					ReadTime:        articles[i].ReadTime,
// 					ImageURL:        articles[i].ImageURL.Ptr(),
// 					ImageCaption:    articles[i].ImageCaption.Ptr(),
// 					Summary:         articles[i].Summary.Ptr(),
// 					Content:         &articles[i].Content,
// 					Tags:            tag,
// 					Video:           video,
// 					VideoAsCover:    articles[i].VideoAsCover,
// 					MetaTitle:       articles[i].MetaTitle.Ptr(),
// 					MetaDescription: articles[i].MetaDescription.Ptr(),
// 					MetaKeywords:    metaKeyword,
// 					Lists:           names,
// 					CreatedAt:       articles[i].CreatedAt,
// 					UpdatedAt:       articles[i].UpdatedAt,
// 				},
// 			})
// 		}
// 	}

// 	res := view.DataResponse{
// 		Type: "lists",
// 		Attributes: view.ListAttributes{
// 			Articles: relatedArticles,
// 		},
// 	}
// 	view.RenderJSONData(w, res, http.StatusOK)
// }

// func (h *handler) handleGetArticleByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var (
// 		project, _ = authpassport.GetProject(r)
// 		pid        = project.ID
// 		_id        = ps.ByName("id")
// 		id, err    = strconv.ParseInt(_id, 10, 64)
// 		resVid     responseVideo
// 	)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	article, err := h.articles.Get(id, pid)
// 	if err == sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Article not found", http.StatusNotFound)
// 		return
// 	}
// 	if err != nil && err != sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get article", http.StatusInternalServerError)
// 		return
// 	}

// 	als, err := h.articleLists.SelectByArticle(id, pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 		return
// 	}

// 	ids := make([]int64, 0, len(als))
// 	for _, al := range als {
// 		ids = append(ids, al.ListID)
// 	}

// 	lists, err := h.lists.SelectByIDs(ids, pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 		return
// 	}

// 	names := make([]string, 0, len(lists))
// 	for _, list := range lists {
// 		names = append(names, list.Name)
// 	}

// 	tag := make([]string, 0, 10)
// 	if article.Tags.Valid && article.Tags.String != "" {
// 		tag = strings.Split(article.Tags.String, ",")
// 	}

// 	metaKeyword := make([]string, 0, 10)
// 	if article.MetaKeywords.Valid && article.MetaKeywords.String != "" {
// 		metaKeyword = strings.Split(article.MetaKeywords.String, ",")
// 	}

// 	var video *videoData
// 	if article.VideoID.Valid && article.VideoID.String != "" {
// 		URL := "https://supersoccer.tv/api/v2/videos/" + article.VideoID.String
// 		request, _ := http.NewRequest("GET", URL, nil)
// 		response, err := h.client.Do(request)
// 		if err == nil && response.StatusCode == 200 {
// 			body, _ := ioutil.ReadAll(response.Body)
// 			_ = json.Unmarshal(body, &resVid)
// 			if len(resVid.Data) > 0 {
// 				video = &resVid.Data[0]
// 			}
// 		}
// 	}

// 	res := view.DataResponse{
// 		Type: "articles",
// 		ID:   article.ID,
// 		Attributes: view.ArticleAttributes{
// 			Title:           article.Title,
// 			Author:          article.Author.Ptr(),
// 			ReadTime:        article.ReadTime,
// 			ImageURL:        article.ImageURL.Ptr(),
// 			ImageCaption:    article.ImageCaption.Ptr(),
// 			Summary:         article.Summary.Ptr(),
// 			Content:         &article.Content,
// 			Tags:            tag,
// 			Video:           video,
// 			VideoAsCover:    article.VideoAsCover,
// 			MetaTitle:       article.MetaTitle.Ptr(),
// 			MetaDescription: article.MetaDescription.Ptr(),
// 			MetaKeywords:    metaKeyword,
// 			Lists:           names,
// 			CreatedAt:       article.CreatedAt,
// 			UpdatedAt:       article.UpdatedAt,
// 		},
// 	}
// 	view.RenderJSONData(w, res, http.StatusOK)
// }

// func (h *handler) handleGetAllArticles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var (
// 		project, _ = authpassport.GetProject(r)
// 		pid        = project.ID
// 		resVid     responseVideo
// 	)

// 	articles, err := h.articles.Select(pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get articles", http.StatusInternalServerError)
// 		return
// 	}

// 	res := make([]view.DataResponse, 0, len(articles))
// 	for _, article := range articles {
// 		als, err := h.articleLists.SelectByArticle(article.ID, pid)
// 		if err != nil {
// 			log.Println(err)
// 			view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 			return
// 		}

// 		ids := make([]int64, 0, len(als))
// 		for _, al := range als {
// 			ids = append(ids, al.ListID)
// 		}

// 		lists, err := h.lists.SelectByIDs(ids, pid)
// 		if err != nil {
// 			log.Println(err)
// 			view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 			return
// 		}

// 		names := make([]string, 0, len(lists))
// 		for _, list := range lists {
// 			names = append(names, list.Name)
// 		}

// 		tag := make([]string, 0, 10)
// 		if article.Tags.Valid && article.Tags.String != "" {
// 			tag = strings.Split(article.Tags.String, ",")
// 		}

// 		metaKeyword := make([]string, 0, 10)
// 		if article.MetaKeywords.Valid && article.MetaKeywords.String != "" {
// 			metaKeyword = strings.Split(article.MetaKeywords.String, ",")
// 		}

// 		var video *videoData
// 		if article.VideoID.Valid && article.VideoID.String != "" {
// 			URL := "https://supersoccer.tv/api/v2/videos/" + article.VideoID.String
// 			request, _ := http.NewRequest("GET", URL, nil)
// 			response, err := h.client.Do(request)
// 			if err == nil && response.StatusCode == 200 {
// 				body, _ := ioutil.ReadAll(response.Body)
// 				_ = json.Unmarshal(body, &resVid)
// 				if len(resVid.Data) > 0 {
// 					video = &resVid.Data[0]
// 				}
// 			}
// 		}

// 		res = append(res, view.DataResponse{
// 			Type: "articles",
// 			ID:   article.ID,
// 			Attributes: view.ArticleAttributes{
// 				Title:           article.Title,
// 				Author:          article.Author.Ptr(),
// 				ReadTime:        article.ReadTime,
// 				ImageURL:        article.ImageURL.Ptr(),
// 				ImageCaption:    article.ImageCaption.Ptr(),
// 				Summary:         article.Summary.Ptr(),
// 				Content:         &article.Content,
// 				Tags:            tag,
// 				Video:           video,
// 				VideoAsCover:    article.VideoAsCover,
// 				MetaTitle:       article.MetaTitle.Ptr(),
// 				MetaDescription: article.MetaDescription.Ptr(),
// 				MetaKeywords:    metaKeyword,
// 				Lists:           names,
// 				CreatedAt:       article.CreatedAt,
// 				UpdatedAt:       article.UpdatedAt,
// 			},
// 		})
// 	}
// 	view.RenderJSONData(w, res, http.StatusOK)
// }

// func (h *handler) handlePostArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var (
// 		project, _ = authpassport.GetProject(r)
// 		pid        = project.ID
// 		params     reqArticle
// 	)

// 	err := form.Bind(&params, r)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	listID := make([]int64, 0, len(params.Lists))
// 	if len(params.Lists) > 0 {
// 		for _, list := range params.Lists {
// 			lists, err := h.lists.GetByName(list, pid)
// 			if err == sql.ErrNoRows {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Lists not found", http.StatusNotFound)
// 				return
// 			}
// 			if err != nil && err != sql.ErrNoRows {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 				return
// 			}
// 			listID = append(listID, lists.ID)
// 		}
// 	}

// 	content := strings.Split(params.Content, " ")
// 	words := len(content)
// 	readTime := int64(words / 265)
// 	if readTime == 0 {
// 		readTime = 1
// 	}
// 	tags := strings.Join(params.Tags, ",")
// 	metaKeywords := strings.Join(params.MetaKeywords, ",")
// 	article := articles.Article{
// 		Title:           params.Title,
// 		Author:          null.StringFromPtr(params.Author),
// 		ReadTime:        readTime,
// 		ImageURL:        null.StringFromPtr(params.ImageURL),
// 		ImageCaption:    null.StringFromPtr(params.ImageCaption),
// 		Summary:         null.StringFromPtr(params.Summary),
// 		Content:         params.Content,
// 		Tags:            null.StringFromPtr(&tags),
// 		VideoID:         null.StringFromPtr(params.VideoID),
// 		VideoAsCover:    params.VideoAsCover,
// 		MetaTitle:       null.StringFromPtr(params.MetaTitle),
// 		MetaDescription: null.StringFromPtr(params.MetaDescription),
// 		MetaKeywords:    null.StringFromPtr(&metaKeywords),
// 		ProjectID:       pid,
// 	}

// 	err = h.articles.Insert(&article)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed post article", http.StatusInternalServerError)
// 		return
// 	}

// 	for _, id := range listID {
// 		err = h.articleLists.Insert(id, article.ID, pid)
// 		if err != nil {
// 			log.Println(err)
// 			view.RenderJSONError(w, "Failed insert to lists", http.StatusInternalServerError)
// 			return
// 		}
// 	}
// 	view.RenderJSONData(w, "OK", http.StatusOK)
// }

// func (h *handler) handlePatchArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var (
// 		project, _ = authpassport.GetProject(r)
// 		pid        = project.ID
// 		params     reqArticle
// 		_id        = ps.ByName("id")
// 		id, err    = strconv.ParseInt(_id, 10, 64)
// 	)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	_, err = h.articles.Get(id, pid)
// 	if err == sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Article not found", http.StatusNotFound)
// 		return
// 	}
// 	if err != nil && err != sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get article", http.StatusInternalServerError)
// 		return
// 	}

// 	err = form.Bind(&params, r)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	listID := make([]int64, 0, len(params.Lists))
// 	if len(params.Lists) > 0 {
// 		for _, list := range params.Lists {
// 			lists, err := h.lists.GetByName(list, pid)
// 			if err == sql.ErrNoRows {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Lists not found", http.StatusNotFound)
// 				return
// 			}
// 			if err != nil && err != sql.ErrNoRows {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
// 				return
// 			}
// 			listID = append(listID, lists.ID)
// 		}
// 	}

// 	content := strings.Split(params.Content, " ")
// 	words := len(content)
// 	readTime := int64(words / 265)
// 	if readTime == 0 {
// 		readTime = 1
// 	}

// 	tags := strings.Join(params.Tags, ",")
// 	metaKeywords := strings.Join(params.MetaKeywords, ",")
// 	article := articles.Article{
// 		ID:              id,
// 		Title:           params.Title,
// 		Author:          null.StringFromPtr(params.Author),
// 		ReadTime:        readTime,
// 		ImageURL:        null.StringFromPtr(params.ImageURL),
// 		ImageCaption:    null.StringFromPtr(params.ImageCaption),
// 		Summary:         null.StringFromPtr(params.Summary),
// 		Content:         params.Content,
// 		Tags:            null.StringFromPtr(&tags),
// 		VideoID:         null.StringFromPtr(params.VideoID),
// 		VideoAsCover:    params.VideoAsCover,
// 		MetaTitle:       null.StringFromPtr(params.MetaTitle),
// 		MetaDescription: null.StringFromPtr(params.MetaDescription),
// 		MetaKeywords:    null.StringFromPtr(&metaKeywords),
// 		ProjectID:       pid,
// 	}

// 	err = h.articles.Update(&article)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed update article", http.StatusInternalServerError)
// 		return
// 	}

// 	err = h.articleLists.Delete(id, pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed delete from lists", http.StatusInternalServerError)
// 		return
// 	}

// 	if len(listID) > 0 {
// 		for _, id := range listID {
// 			err = h.articleLists.Insert(id, article.ID, pid)
// 			if err != nil {
// 				log.Println(err)
// 				view.RenderJSONError(w, "Failed insert to lists", http.StatusInternalServerError)
// 				return
// 			}
// 		}
// 	}
// 	view.RenderJSONData(w, "OK", http.StatusOK)
// }

// func (h *handler) handleDeleteArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var (
// 		project, _ = authpassport.GetProject(r)
// 		pid        = project.ID
// 		_id        = ps.ByName("id")
// 		id, err    = strconv.ParseInt(_id, 10, 64)
// 	)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
// 		return
// 	}

// 	_, err = h.articles.Get(id, pid)
// 	if err == sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Article not found", http.StatusNotFound)
// 		return
// 	}
// 	if err != nil && err != sql.ErrNoRows {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed get article", http.StatusInternalServerError)
// 		return
// 	}

// 	err = h.articles.Delete(id, pid)
// 	if err != nil {
// 		log.Println(err)
// 		view.RenderJSONError(w, "Failed delete article", http.StatusInternalServerError)
// 		return
// 	}

// 	view.RenderJSONData(w, "OK", http.StatusOK)
// }
