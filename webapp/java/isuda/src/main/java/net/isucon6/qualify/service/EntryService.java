package net.isucon6.qualify.service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import net.isucon6.qualify.dto.EntryDto;
import net.isucon6.qualify.mapper.EntryMapper;
import org.modelmapper.ModelMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class EntryService {
    @Autowired
    private EntryMapper entryMapper;
    @Autowired
    private ModelMapper modelMapper;

    public List<EntryDto> findEntries(int perPage, int currentPage) {
        Map<String, Integer> params = new HashMap<>();
        params.put("perPage", perPage);
        params.put("offset", (perPage * (currentPage - 1)));
        List<EntryDto> entries = entryMapper.findByPageNum(params).stream()
                .map(e -> {
                    EntryDto ed = modelMapper.map(e, EntryDto.class);
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
