package techaas.max_uni.uni_back.controller;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import techaas.max_uni.uni_back.dao.entity.News;
import techaas.max_uni.uni_back.dao.repository.NewsRepository;

import java.util.List;

@RestController
@RequiredArgsConstructor
public class NewsController {

    private final NewsRepository newsRepository;

    @GetMapping("/news")
    public ResponseEntity<List<News>> getLastNews() {
        return ResponseEntity.ok(newsRepository.findLast10News(10));
    }

    @PostMapping("/news/save")
    public ResponseEntity<?> saveNews(@RequestBody News news) {
        newsRepository.save(news);
        return ResponseEntity.noContent().build();
    }

    @PatchMapping("/news/{newsId}")
    public ResponseEntity<?> editNews(@PathVariable Long newsId, @RequestBody String text) {
        newsRepository.editText(text, newsId);
        return ResponseEntity.noContent().build();
    }
}
