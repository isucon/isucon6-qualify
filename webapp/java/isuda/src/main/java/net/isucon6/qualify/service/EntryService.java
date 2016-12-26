package net.isucon6.qualify.service;

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
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

@Service
public class EntryService {
    @Autowired
    private EntryMapper entryMapper;
    @Autowired
    private ModelMapper modelMapper;

    private String htmlify(final String content) {
        if (StringUtils.isEmpty(content)) {
            return "";
        }

        List<Entry> keywords = entryMapper.findAllOrderByLength();

        Pattern regexp = Pattern.compile(keywords.stream()
                .map(Entry::getKeyword)
                .map(Pattern::quote)
                .collect(Collectors.joining("|")));
        Matcher matcher = regexp.matcher(content);
        Map<String, String> kw2sha = keywords.stream()
                .collect(Collectors.toMap(
                        Entry::getKeyword,
                        k -> DigestUtils.sha1Hex(("isuda_" + k.getKeyword()).getBytes())
                ));

        StringBuffer sbKw2Sha = new StringBuffer();
        while (matcher.find()) {
            matcher.appendReplacement(sbKw2Sha, kw2sha.get(matcher.group()));
        }
        String result = matcher.appendTail(sbKw2Sha).toString();

        for (Map.Entry<String, String> e : kw2sha.entrySet()) {
            String kw = e.getKey();
            String hash = e.getValue();
            String link = String.format("<a href=\"%s\">%s</a>",
                    String.format("/keyword/%s", kw),
                    kw /* TODO: HTML Escape kw */
            );
            Matcher m = Pattern.compile(hash).matcher(result);
            result = m.replaceAll(link);
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
