package net.isucon6.qualify.service;

import java.io.UnsupportedEncodingException;
import java.net.URLEncoder;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

import net.isucon6.qualify.domain.Entry;
import net.isucon6.qualify.dto.EntryDto;
import net.isucon6.qualify.mapper.EntryMapper;
import org.apache.commons.codec.digest.DigestUtils;
import org.modelmapper.ModelMapper;
import org.slf4j.Logger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;
import org.springframework.web.util.HtmlUtils;

@Service
public class EntryService {
    private final EntryMapper entryMapper;
    private final ModelMapper modelMapper;
    private final Logger log = org.slf4j.LoggerFactory.getLogger(EntryService.class);

    @Autowired
    public EntryService(EntryMapper entryMapper, ModelMapper modelMapper) {
        this.entryMapper = entryMapper;
        this.modelMapper = modelMapper;
    }

    private String htmlify(final String content) {
        if (StringUtils.isEmpty(content)) {
            return "";
        }

        List<Entry> keywords = entryMapper.findAllOrderByLength();

        Matcher matcher = Pattern.compile(keywords.stream()
                .map(Entry::getKeyword)
                .map(Pattern::quote)
                .collect(Collectors.joining("|", "(", ")"))).matcher(content);
        Map<String, String> kw2sha = keywords.stream()
                .collect(Collectors.toMap(
                        Entry::getKeyword,
                        k -> "isuda_" + DigestUtils.sha1Hex(k.getKeyword())
                ));
        StringBuffer sbKw2Sha = new StringBuffer();
        while (matcher.find()) {
            matcher.appendReplacement(sbKw2Sha, kw2sha.get(matcher.group()));
        }
        String result = matcher.appendTail(sbKw2Sha).toString();
        try {
            for (Map.Entry<String, String> e : kw2sha.entrySet()) {
                String kw = e.getKey();
                String hash = e.getValue();
                String link = String.format("<a href=\"%s\">%s</a>",
                        String.format("/keyword/%s", URLEncoder.encode(kw, "UTF-8")),
                        HtmlUtils.htmlEscape(kw, "UTF-8")
                );
                Matcher m = Pattern.compile(hash).matcher(result);
                result = m.replaceAll(link);
            }
        } catch (UnsupportedEncodingException e) {
            log.warn("Failed to replace keyword.");
        }

        return result.replace("\n", "<br />");
    }

    public List<EntryDto> findEntries(int perPage, int currentPage) {
        Map<String, Integer> params = new HashMap<>();
        params.put("perPage", perPage);
        params.put("offset", (perPage * (currentPage - 1)));
        List<EntryDto> entries = entryMapper.findByPageNum(params).stream()
                .map(e -> {
                    EntryDto ed = modelMapper.map(e, EntryDto.class);
                    ed.setHtml(htmlify(e.getDescription()));
                    // TODO: starのAPIを毎回叩く
                    ed.setStars(new ArrayList<String>() {{
                        add("blackawa");
                    }});
                    return ed;
                })
                .collect(Collectors.toList());
        return entries;
    }
}